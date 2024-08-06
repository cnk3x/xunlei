package vms

import (
	"context"
	"log/slog"

	"github.com/cnk3x/xunlei/pkg/lod"
	"github.com/cnk3x/xunlei/pkg/mounts"
)

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

func chroot(ctx context.Context, target string) (err error) {
	return mounts.Chroot(ctx, target)
}
