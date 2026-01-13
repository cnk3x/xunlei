package rootfs

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/fo"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/samber/lo"
)

type Undo func()
type Option func(ro *RunOptions)

type RunOptions struct {
	root   string
	mounts []MountOptions
	links  []LinkOptions
	// force  bool

	before func(ctx context.Context) error
	after  func() error
}

type MountOptions struct {
	Target   string
	Source   []MountSourceOptions
	Bind     bool
	Optional bool
}

type MountSourceOptions struct {
	Source string
	Fstype string
	Flags  uintptr
	Data   string
}

type LinkOptions struct {
	Target   string
	Source   []string
	DirMode  fs.FileMode
	Optional bool
}

func Run(ctx context.Context, newRoot string, run func(ctx context.Context) error, options ...Option) (err error) {
	ctx = log.Prefix(ctx, "boot")

	slog.InfoContext(ctx, "start boot")

	defer func() {
		if errors.Is(err, context.Canceled) {
			slog.InfoContext(ctx, "rootfs canceled")
			err = nil
		}
	}()

	var opts RunOptions
	if opts.root, err = filepath.Abs(newRoot); err != nil {
		return
	}

	for _, option := range options {
		option(&opts)
	}

	if opts.root == "" || opts.root == "/" {
		err = run(ctx)
		return
	}

	// //兜底
	// defer func() {
	// 	if opts.force {
	// 		exec.Command("sh", "-c", `mount | grep `+newRoot+` | grep on | awk '{print $3}' | xargs -I {} umount {}`).Run()
	// 	}
	// }()

	//undos
	var undos []Undo
	defer execUndo(makeUndos(&undos), nil)

	var undo Undo

	//mkdirs
	dirs := lo.Map(opts.mounts, func(m MountOptions, _ int) string { return m.Target })
	if undo, err = Mkdirs(ctx, dirs, 0777, true); err != nil {
		return
	}
	undos = append(undos, undo)

	//mounts
	if undo, err = mounts(ctx, opts.mounts); err != nil {
		return
	}
	undos = append(undos, undo)

	//links
	if undo, err = links(ctx, opts.links); err != nil {
		return
	}
	undos = append(undos, undo)

	//after
	defer func() {
		if opts.after != nil {
			if err = opts.after(); err != nil {
				return
			}
		}
	}()

	//before
	if opts.before != nil {
		if err = opts.before(ctx); err != nil {
			return
		}
	}

	//chroot & run
	err = chroot(log.Prefix(ctx, "prog"), opts.root, run)
	return
}

// chroot & run
func chroot(ctx context.Context, newRoot string, run func(ctx context.Context) error) (err error) {
	wd, e := os.Getwd()
	if err = e; err != nil {
		return
	}

	var rfd int
	if rfd, err = syscall.Open("/", syscall.O_RDONLY, 0); err != nil {
		return
	}

	defer func() {
		logWithErr(ctx, syscall.Close(rfd), "closeFd", "fd", rfd)
	}()

	if err = logWithErr(ctx, syscall.Chdir(newRoot), "chdir", "dir", newRoot); err != nil {
		return
	}

	if err = logWithErr(ctx, syscall.Chroot("."), "chroot", "path", "."); err != nil {
		return
	}

	defer func() {
		logWithErr(ctx, syscall.Fchdir(rfd), "fchdir", "fd", rfd)
		logWithErr(ctx, syscall.Chroot("."), "chroot rollback", "path", ".")
		logWithErr(ctx, syscall.Chdir(wd), "chdir", "dir", wd)
	}()

	if err = logWithErr(ctx, syscall.Chdir("/"), "chdir", "dir", "/"); err != nil {
		return
	}

	//run
	err = run(ctx)
	return
}

func Mkdirs(ctx context.Context, dirs []string, perm fs.FileMode, existsOk bool) (undo Undo, err error) {
	var undos []Undo
	undo = makeUndos(&undos)
	defer execUndo(undo, &err)
	for _, dir := range dirs {
		u, e := Mkdir(ctx, dir, perm, existsOk)
		if e != nil {
			err = e
			return
		}
		undos = append(undos, u)
	}
	return
}

// 脱了裤子放个屁，为了能够方便回滚
func Mkdir(ctx context.Context, dir string, perm fs.FileMode, existsOk bool) (undo Undo, err error) {
	if dir, err = filepath.Abs(dir); err != nil {
		return
	}

	var undos []Undo
	undo = makeUndos(&undos)
	defer execUndo(undo, &err)

	vol := filepath.VolumeName(dir)
	if vol != "" {
		dir = dir[len(vol):]
	}
	dir = strings.Trim(dir, Separator)

	fields := strings.FieldsFunc(dir, func(r rune) bool { return r == filepath.Separator })
	for i := range fields {
		full := filepath.Join(cmp.Or(vol, Separator), filepath.Join(fields[:i+1]...))
		switch stat, e := os.Stat(full); {
		case e != nil && !os.IsNotExist(e):
			err = e
		case e != nil:
			if err = os.Mkdir(full, perm); err != nil {
				err = fmt.Errorf("mkdir fail: %w: %s", err, full)
			} else {
				slog.DebugContext(ctx, "mkdir", "path", full)
				undos = append(undos, func() { logWithErr(ctx, os.Remove(full), "rmdir", "path", full) })
			}
		case !stat.IsDir():
			err = fmt.Errorf("mkdir fail, parent is not a directory: %w: %s", fs.ErrExist, full)
		case !existsOk:
			err = fmt.Errorf("mkdir fail, parent is exists: %w: %s", fs.ErrExist, full)
		}

		if err != nil {
			slog.ErrorContext(ctx, "mkdir", "path", full, "err", err)
			return
		}
	}

	return
}

func mounts(ctx context.Context, mountPoints []MountOptions) (undo Undo, err error) {
	var undos []Undo
	undo = makeUndos(&undos)
	defer execUndo(undo, &err)

	for _, m := range mountPoints {
		r, e := mount(ctx, m)
		if e != nil {
			err = e
			return
		}
		undos = append(undos, r)
	}
	return
}

func mount(ctx context.Context, m MountOptions) (undo Undo, err error) {
	defer func() {
		if err != nil {
			slog.DebugContext(ctx, string(utils.Eon(json.Marshal(m))))
		}
	}()

	defer execUndo(undo, &err)

	if len(m.Source) == 0 {
		slog.WarnContext(ctx, "mount", "target", m.Target, "optional", m.Optional, "err", "source not provided")
		if !m.Optional {
			err = fmt.Errorf("source not provided")
		}
		return
	}

	for _, mm := range m.Source {
		src := mm.Source

		flag := mm.Flags
		if m.Bind {
			if src, err = filepath.EvalSymlinks(src); err != nil {
				continue
			}
			if flag == 0 {
				flag = syscall.MS_BIND
				// | syscall.MS_REC, //todo: 是否需要递归绑定?
			}
		}

		if err = syscall.Mount(src, m.Target, mm.Fstype, flag, mm.Data); err != nil {
			slog.WarnContext(ctx, "mount", "target", m.Target, "source", src, "optional", m.Optional, "err", err)
			continue
		}

		unmount := func() error { return syscall.Unmount(m.Target, syscall.MNT_DETACH|syscall.MNT_FORCE) }
		undo = func() { logWithErr(ctx, unmount(), "unmount", "target", m.Target) }

		slog.DebugContext(ctx, "mount", "target", m.Target, "source", src, "optional", m.Optional)
		break
	}

	return
}

func links(ctx context.Context, links []LinkOptions) (undo Undo, err error) {
	var undos []Undo
	undo = makeUndos(&undos)
	defer execUndo(undo, &err)

	for _, l := range links {
		for _, source := range l.Source {
			r, e := link(ctx, source, l.Target, l.DirMode)
			if e != nil {
				err = e
				return
			}
			undos = append(undos, r)
		}
	}

	return
}

// link hard link, only for file
func link(ctx context.Context, source string, linkFile string, dirMode fs.FileMode) (undo Undo, err error) {
	var undos []Undo
	undo = makeUndos(&undos)
	defer execUndo(undo, &err)

	var realPath string
	if realPath, err = filepath.EvalSymlinks(source); err != nil {
		return
	}

	var dirUndo Undo
	if dirUndo, err = Mkdir(ctx, filepath.Dir(linkFile), dirMode, true); err != nil {
		return
	}
	undos = append(undos, dirUndo)

	op := "link"
	if err = os.Link(realPath, linkFile); errors.Is(err, syscall.EXDEV) {
		op, err = "copy", fileCopy(realPath, linkFile)
	}

	logWithErr(ctx, err, "link", "op", op, "link", linkFile, "source", source, "realpath", realPath)
	undos = append(undos, func() { logWithErr(ctx, os.Remove(linkFile), "unlink", "link", linkFile) })
	return
}

func fileCopy(source, target string) error {
	return fo.OpenRead(source, func(src *os.File) (err error) { return fo.OpenWrite(target, fo.From(src), fo.PermFrom(src)) })
}

func iif[T any](c bool, t, f T) T {
	if c {
		return t
	}
	return f
}

func execUndo(undo Undo, err *error) {
	if err == nil || *err != nil && undo != nil {
		undo()
	}
}

func makeUndos(undos *[]Undo) (undo Undo) {
	return func() {
		if undos == nil || len(*undos) == 0 {
			return
		}
		for _, undo := range slices.Backward(*undos) {
			if undo != nil {
				undo()
			}
		}
	}
}

func logWithErr(ctx context.Context, err error, msg string, args ...any) error {
	switch {
	case errors.Is(err, syscall.ENOTEMPTY):
		slog.Log(ctx, slog.LevelDebug, msg, append(args, "err", "skip because not empty")...)
	case errors.Is(err, context.Canceled):
		slog.Log(ctx, slog.LevelDebug, msg, append(args, "err", context.Cause(ctx))...)
	case err != nil:
		slog.Log(ctx, slog.LevelWarn, msg, append(args, "err", err)...)
	default:
		slog.Log(ctx, slog.LevelDebug, msg, args...)
	}
	return err
}
