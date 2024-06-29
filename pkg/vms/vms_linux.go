package vms

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cnk3x/xunlei/pkg/cmd"
	"github.com/cnk3x/xunlei/pkg/lod"
	"github.com/cnk3x/xunlei/pkg/mounts"
)

const (
	TAG_MAIN   = "main"
	TAG_CHROOT = "chroot"
)

type Vm struct {
	Root    string
	Binds   []string
	Args    []string
	Env     cmd.EnvSet
	Uid     string
	Gid     string
	RootRun func(ctx context.Context) error
	UserRun func(ctx context.Context, uid, gid uint32) error
	Main    func(ctx context.Context) error
}

func (d *Vm) Run(ctx context.Context) (err error) {
	switch {
	case cmd.IsForkTag(TAG_MAIN):
		err = d.Main(ctx)
	case cmd.IsForkTag(TAG_CHROOT):
		err = d.chrootFork(ctx)
	case d.Root != "" && d.Root != "/":
		err = d.mountFork(ctx)
	default:
		if d.RootRun != nil {
			if err = d.RootRun(ctx); err != nil {
				return
			}
		}

		var uid, gid uint32
		if uid, gid, err = cmd.UserFind(d.Uid, d.Gid); err != nil {
			err = fmt.Errorf("lookup uid/gid fail: %w", err)
			return
		}

		if uid != 0 || gid != 0 && d.UserRun != nil {
			if err = d.UserRun(ctx, uid, gid); err != nil {
				return
			}
		}

		err = d.Main(ctx)
	}
	return
}

func (d *Vm) chrootFork(ctx context.Context) (err error) {
	if err = mounts.Chroot(ctx, d.Root); err != nil {
		return
	}

	if d.RootRun != nil {
		if err = d.RootRun(ctx); err != nil {
			return
		}
	}

	var uid, gid uint32
	if uid, gid, err = cmd.UserFind(d.Uid, d.Gid); err != nil {
		err = fmt.Errorf("lookup uid/gid fail: %w", err)
		return
	}

	if uid != 0 || gid != 0 && d.UserRun != nil {
		if err = d.UserRun(ctx, uid, gid); err != nil {
			return
		}
	}

	err = cmd.Fork(ctx, cmd.ForkOptions{Tag: TAG_MAIN, Args: d.Args, Env: d.Env, Uid: uid, Gid: gid})
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
		"/bin", "/dev",
		"/lib", "/lib64",
		"/tmp", "/run", "/proc",
		"/etc/hosts", "/etc/hostname",
		"/etc/timezone", "/etc/localtime",
		"/etc/resolv.conf", "/etc/profile",
	)

	if err = mounts.Binds(ctx, d.Root, endpoints, true); err != nil {
		return
	}

	return
}
