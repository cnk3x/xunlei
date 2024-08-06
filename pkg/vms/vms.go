package vms

import (
	"context"
	"log/slog"
	"os"

	"github.com/cnk3x/xunlei/pkg/cmd"
	"github.com/cnk3x/xunlei/pkg/log"
)

type Vm struct {
	Root   string
	Binds  []string
	Args   []string
	Env    cmd.EnvSet
	Uid    uint32
	Gid    uint32
	PreRun func(ctx context.Context) error
	Main   func(ctx context.Context) error
}

const (
	TAG_CHROOT = "chroot"
)

func (d *Vm) Run(ctx context.Context) (err error) {
	if d.Root == "" {
		d.Root = "/"
	}
	switch {
	case cmd.IsForkTag(TAG_CHROOT):
		err = d.chrootFork(ctx)
	case d.Root != "/":
		if cmd.IsForkTag(d.Root) {
			ctx = log.Prefix(ctx, d.Root)
			slog.InfoContext(ctx, "check current user", "uid", os.Getuid(), "gid", os.Getgid())
			err = d.Main(ctx)
		} else {
			err = d.mountFork(ctx)
		}
	default:
		err = d.directRun(ctx)
	}
	return
}

func (d *Vm) directRun(ctx context.Context) (err error) {
	if d.PreRun != nil {
		if err = d.PreRun(ctx); err != nil {
			return
		}
	}

	slog.InfoContext(ctx, "check current user", "uid", os.Getuid(), "gid", os.Getgid())
	err = d.Main(ctx)
	return
}

func (d *Vm) chrootFork(ctx context.Context) (err error) {
	if err = chroot(ctx, d.Root); err != nil {
		return
	}

	ctx = log.Prefix(ctx, d.Root)

	if d.PreRun != nil {
		if err = d.PreRun(ctx); err != nil {
			return
		}
	}

	err = cmd.Fork(ctx, cmd.ForkOptions{Tag: d.Root, Args: d.Args, Env: d.Env, Uid: d.Uid, Gid: d.Gid})
	return
}

func (d *Vm) mountFork(ctx context.Context) (err error) {
	defer d.unmount(ctx)
	if err = d.mount(ctx); err != nil {
		return
	}
	return cmd.Fork(ctx, cmd.ForkOptions{Tag: TAG_CHROOT, Args: d.Args, Env: d.Env})
}
