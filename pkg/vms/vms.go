package vms

import (
	"context"
	"log/slog"

	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/cnk3x/xunlei/pkg/vms/sys"
)

// type Undo func()
type Option func(ro *options)

type options struct {
	root string

	uid, gid int

	mounts []sys.MountOptions
	binds  []sys.BindOptions
	links  []sys.LinkOptions

	before func(ctx context.Context) (undo func(), err error)
	after  func(ctx context.Context, err error) error
	run    func(ctx context.Context) error

	debug bool
}

// Exec 通过哦 chroot 实现文件系统隔离
//
//   - chroot + seteuid 实现权限最小化（临时降权）
//   - 执行顺序 root 启动 → 准备 chroot 监狱 → chroot 切换 → setegid/seteuid 降权 → 执行核心任务 → 恢复 root 权限
func Exec(ctx context.Context, execOpts ...Option) (err error) {
	slog.InfoContext(ctx, "vms start")
	defer slog.InfoContext(ctx, "vms done")

	var opts options
	for _, option := range execOpts {
		option(&opts)
	}

	undo, e := utils.SeqExecWithUndo(
		func() (sys.Undo, error) { return sys.Mounts(ctx, opts.mounts) },           //mounts
		func() (sys.Undo, error) { return sys.BindRs(ctx, opts.root, opts.binds) }, //binds
		func() (sys.Undo, error) { return sys.LinkRs(ctx, opts.root, opts.links) }, //links
	)
	if err = e; err != nil {
		return
	}
	defer undo()

	if opts.before != nil {
		var beforeUndo sys.Undo
		if beforeUndo, err = opts.before(ctx); err != nil {
			return
		}
		defer beforeUndo()
	}

	run := func(ctx context.Context) error {
		slog.InfoContext(ctx, "runner start")
		defer slog.InfoContext(ctx, "runner done")
		return sys.RunAs(ctx, opts.uid, opts.gid, func() error {
			if opts.run == nil {
				return nil
			}
			return opts.run(ctx)
		})
	}

	if opts.root == "" || opts.root == "/" {
		err = run(ctx)
	} else {
		err = sys.Chroot(ctx, opts.root, run, opts.debug) //chroot & run
	}

	if opts.after != nil {
		if err = opts.after(ctx, err); err != nil {
			return
		}
	}
	return
}
