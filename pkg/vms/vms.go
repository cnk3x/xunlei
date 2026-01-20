package vms

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"syscall"

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
	check := func(name string, err error, args ...any) error {
		if err == nil {
			slog.DebugContext(ctx, "call "+name, args...)
			// } else {
			// 	slog.DebugContext(ctx, "call "+name, append(args, "err", err)...)
		}
		return err
	}

	checkUid := func(name string) (err error) {
		if syscall.Getuid() != 0 {
			err = fmt.Errorf("%s: only the root process (UID=0) supports.", name)
		}
		return
	}

	if root != "" && root != "/" {
		if err = check("check root", checkUid("chroot")); err != nil {
			return
		}

		wd, e := os.Getwd()
		if err = check("getwd", e, "wd", wd); err != nil {
			return
		}

		fd, e := syscall.Open("/", syscall.O_RDONLY, 0)
		if err = check("openFd", e, "dir", "/", "fd", fd); err != nil {
			return
		}
		defer func() { check("closeFd", syscall.Close(fd), "fd", fd) }()

		if err = check("chdir", os.Chdir(root), "dir", root); err != nil {
			return
		}
		defer func() { check("chdir", os.Chdir(wd), "dir", wd) }()

		if err = check("chroot", syscall.Chroot("."), "dir", "."); err != nil {
			return
		}
		defer func() { check("chroot", syscall.Chroot("."), "fd", fd) }()
		defer func() { check("fchdir", syscall.Fchdir(fd), "fd", fd) }()

		if err = check("chdir", os.Chdir("/"), "dir", "/"); err != nil {
			return
		}
	}

	if opts.uid > 0 || opts.gid > 0 {
		if err = check("check root", checkUid("chroot")); err != nil {
			return
		}

		if opts.gid > 0 {
			if err = check("setegid", syscall.Setegid(opts.gid), "gid", opts.gid); err != nil {
				return
			}
			defer func() { check("setegid", syscall.Setegid(0), "gid", 0) }()
		}

		if opts.uid > 0 {
			if err = check("seteuid", syscall.Seteuid(opts.uid), "uid", opts.uid); err != nil {
				return
			}
			defer func() { check("seteuid", syscall.Seteuid(0), "uid", 0) }()
		}
	}

	err = run(ctx)
	return
}
