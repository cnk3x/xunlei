package vms

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/cnk3x/xunlei/pkg/vms/sys"
)

// type Undo func()
type Option func(ro *options)

type options struct {
	root string
	uid  int
	gid  int
	wait bool

	mounts  []sys.MountOptions
	binds   []sys.BindOptions
	links   []sys.LinkOptions
	symbols []sys.LinkOptions

	before func(ctx context.Context) (undo func(), err error)
	after  func(ctx context.Context, err error) error
	run    func(ctx context.Context) error
}

// Execute 通过哦 chroot 实现文件系统隔离
//
//   - chroot + seteuid 实现权限最小化（临时降权）
//   - 执行顺序 root 启动 → 准备 chroot 监狱 → chroot 切换 → setegid/seteuid 降权 → 执行核心任务 → 恢复 root 权限
func Execute(ctx context.Context, execOpts ...Option) (err error) {
	var opts options
	for _, option := range execOpts {
		option(&opts)
	}

	if opts.root != "" && opts.root != "/" {
		//mounts
		unmounts, e := sys.Mounts(ctx, opts.mounts)
		if err = e; err != nil {
			return
		}
		defer unmounts()

		//binds
		unbinds, e := sys.BindRs(ctx, opts.root, opts.binds)
		if err = e; err != nil {
			return
		}
		defer unbinds()

		//links
		unlinks, e := sys.LinkRs(ctx, opts.root, opts.links)
		if err = e; err != nil {
			return
		}
		defer unlinks()

		unSymlinks, e := sys.Symlinks(ctx, opts.symbols)
		if err = e; err != nil {
			return
		}
		defer unSymlinks()
	}

	//before
	if opts.before != nil {
		beforeRestore, e := opts.before(ctx)
		if err = e; err != nil {
			return
		}
		defer beforeRestore()
	}

	//chroot
	err = chrootRun(ctx, opts.root, opts.run, opts)

	if err != nil && opts.wait {
		slog.ErrorContext(ctx, "chroot run", "err", err.Error())
		slog.WarnContext(ctx, "调试模式，在有错误发生的情况下，会一直不退出，等待人工调试，请按 ctrl+c 退出.")
		<-ctx.Done()
	}

	//after
	if opts.after != nil {
		if err = opts.after(ctx, err); err != nil {
			return
		}
	}
	return
}

// chroot & run
func chrootRun(ctx context.Context, root string, run func(ctx context.Context) error, opts options) (err error) {
	chRootBack, e := chroot(ctx, root)
	if err = e; err != nil {
		return
	}
	defer chRootBack()

	chUserBack, e := chuser(ctx, opts.uid, opts.gid)
	if err = e; err != nil {
		return
	}
	defer chUserBack()

	err = run(ctx)
	return
}

func chroot(ctx context.Context, root string) (rollback func(), err error) {
	slog.InfoContext(ctx, "chroot start", "root", root)
	defer slog.InfoContext(ctx, "chroot done")

	bq := utils.BackQueue(&rollback, &err)
	defer bq.ErrDefer()

	bq.Put(func() { slog.InfoContext(ctx, "chroot rollback done") })
	defer bq.Put(func() { slog.InfoContext(ctx, "chroot rollback start") })

	if root != "" && root != "/" {
		if err = errcheck(ctx, slog.LevelDebug, "check root", checkUid("chroot")); err != nil {
			return
		}

		wd, e := os.Getwd()
		if err = errcheck(ctx, slog.LevelDebug, "getwd", e, "wd", wd); err != nil {
			return
		}

		fd, e := syscall.Open("/", syscall.O_RDONLY, 0)
		if err = errcheck(ctx, slog.LevelDebug, "openFd", e, "dir", "/", "fd", fd); err != nil {
			return
		}
		bq.Put(func() { errcheck(ctx, slog.LevelDebug, "closeFd", syscall.Close(fd), "fd", fd) })

		if err = errcheck(ctx, slog.LevelDebug, "chdir", os.Chdir(root), "dir", root); err != nil {
			return
		}
		bq.Put(func() { errcheck(ctx, slog.LevelDebug, "chdir", os.Chdir(wd), "dir", wd) })

		if err = errcheck(ctx, slog.LevelDebug, "chroot", syscall.Chroot("."), "dir", "."); err != nil {
			return
		}
		bq.Put(func() { errcheck(ctx, slog.LevelDebug, "chroot", syscall.Chroot("."), "fd", fd) })
		bq.Put(func() { errcheck(ctx, slog.LevelDebug, "fchdir", syscall.Fchdir(fd), "fd", fd) })

		if err = errcheck(ctx, slog.LevelDebug, "chdir", os.Chdir("/"), "dir", "/"); err != nil {
			return
		}
	}
	return
}

func chuser(ctx context.Context, uid, gid int) (rollback func(), err error) {
	slog.InfoContext(ctx, "chuser start", "uid", uid, "gid", gid)
	defer slog.InfoContext(ctx, "chuser done")

	bq := utils.BackQueue(&rollback, &err)
	defer bq.ErrDefer()

	bq.Put(func() { slog.InfoContext(ctx, "chuser rollback done") })
	defer bq.Put(func() { slog.InfoContext(ctx, "chuser rollback start") })

	if uid > 0 || gid > 0 {
		if err = errcheck(ctx, slog.LevelDebug, "check root", checkUid("chroot")); err != nil {
			return
		}

		if gid > 0 {
			if err = errcheck(ctx, slog.LevelInfo, "setegid", syscall.Setegid(gid), "gid", gid); err != nil {
				return
			}
			bq.Put(func() { errcheck(ctx, slog.LevelInfo, "setegid", syscall.Setegid(0), "gid", 0) })
		}

		if uid > 0 {
			if err = errcheck(ctx, slog.LevelInfo, "seteuid", syscall.Seteuid(uid), "uid", uid); err != nil {
				return
			}
			bq.Put(func() { errcheck(ctx, slog.LevelInfo, "seteuid", syscall.Seteuid(0), "uid", 0) })
		}
	}
	return
}

func errcheck(ctx context.Context, level slog.Level, name string, err error, args ...any) error {
	if err == nil {
		slog.Log(ctx, level, "call "+name, args...)
	} else {
		slog.Log(ctx, utils.Iif(level < 0, level, slog.LevelWarn), "call "+name, append(args, "err", err.Error())...)
	}
	return err
}

func checkUid(name string) (err error) {
	if syscall.Getuid() != 0 {
		err = fmt.Errorf("%s: only the root process (UID=0) supports.", name)
	}
	return
}
