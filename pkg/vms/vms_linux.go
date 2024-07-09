package vms

import (
	"context"
	"github.com/cnk3x/xunlei/pkg/cmd"
	"github.com/cnk3x/xunlei/pkg/lod"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/mounts"
	"log/slog"
	"os"
)

//goland:noinspection GoSnakeCaseUsage
const (
	TAG_CHROOT = "chroot"
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
	if err = mounts.Chroot(ctx, d.Root); err != nil {
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

func (d *Vm) unmount(ctx context.Context) {
	select {
	case <-ctx.Done():
		ctx = context.Background()
	default:
	}

	err := mounts.Unbind(ctx, d.Root)
	slog.Log(ctx, lod.ErrDebug(err), "unmount", "root", d.Root, "err", err)
}

func (d *Vm) mount(ctx context.Context) (err error) {
	slog.InfoContext(ctx, "mount...")

	//如果源目录不存在则不会绑定
	endpoints := append(
		d.Binds,
		"/proc",
		"/bin", "/dev",
		"/lib", "/lib64",
		"/tmp", "/run",
		"/etc/hosts", "/etc/hostname",
		"/etc/timezone", "/etc/localtime",
		"/etc/resolv.conf", "/etc/profile",
	)

	_, endpoints, _, err = ResolvePath(d.Root, endpoints...)

	if err = mounts.Binds(ctx, d.Root, endpoints, true); err != nil {
		return
	}

	return
}
