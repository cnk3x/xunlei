package vms

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/vms/sys"
)

// type Undo func()
type Option func(ro *options)

type options struct {
	root   string
	uid    int
	gid    int
	wait   bool
	mounts []sys.MountOptions
	binds  []sys.BindOptions
	links  []sys.LinkOptions
	before func(ctx context.Context) (undo func(), err error)
	after  func(ctx context.Context, err error) error
	run    func(ctx context.Context) error
}

// Execute 通过哦 chroot 实现文件系统隔离
//
//   - chroot + seteuid 实现权限最小化（临时降权）
//   - 执行顺序 root 启动 → 准备 chroot 监狱 → chroot 切换 → setegid/seteuid 降权 → 执行核心任务 → 恢复 root 权限
func Execute(ctx context.Context, execOpts ...Option) (err error) {
	defer log.LogDone(ctx, slog.LevelInfo, "vms", &err).Defer()

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
	err = chrootRun(ctx, opts.root,
		func() error {
			defer log.LogDone(ctx, slog.LevelInfo, "runner", &err).Defer()
			return opts.run(ctx)
		},
		opts,
	)

	if opts.after != nil {
		if err = opts.after(ctx, err); err != nil {
			return
		}
	}
	return
}

// chroot & run
func chrootRun(ctx context.Context, root string, run func() error, options options) (err error) {
	check := func(name string, err error, args ...any) error {
		if err != nil {
			slog.DebugContext(ctx, "call "+name, append(args, "err", err)...)
		} else {
			slog.DebugContext(ctx, "call "+name, args...)
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
		if err = check("chdir", os.Chdir("/"), "dir", "/"); err != nil {
			return
		}
		defer func() { check("fchdir", syscall.Fchdir(fd), "fd", fd) }()
	}

	if oEUid, oEGid := syscall.Geteuid(), syscall.Getegid(); options.uid != oEUid || options.gid != oEGid {
		if err = check("check root", checkUid("chroot")); err != nil {
			return
		}

		if options.gid != oEGid {
			if err = check("setegid", syscall.Setegid(options.gid), "gid", options.gid); err != nil {
				return
			}
			defer func() { check("setegid", syscall.Setegid(oEGid), "gid", oEGid) }()
		}

		if options.uid != oEUid {
			if err = check("seteuid", syscall.Setegid(options.uid), "uid", options.uid); err != nil {
				return
			}
			defer func() { check("seteuid", syscall.Setegid(oEUid), "uid", oEUid) }()
		}
	}

	if err = check("run", run()); err != nil && options.wait {
		<-ctx.Done()
	}
	return
}
