package sys

import (
	"context"
	"log/slog"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/utils"
)

type MountOptions struct {
	Target   string
	Source   string
	Fstype   string
	Flags    uintptr
	Data     string
	Optional bool
	Root     string
}

func Mounts(ctx context.Context, mounts []MountOptions) (undo Undo, err error) {
	return doMulti(ctx, mounts, Mount)
}

func MountTask(ctx context.Context, mounts ...MountOptions) func() (undo Undo, err error) {
	return func() (undo Undo, err error) { return Mounts(ctx, mounts) }
}

// 完整的绑定
func Mount(ctx context.Context, m MountOptions) (undo Undo, err error) {
	bq := utils.BackQueue(&undo, &err)
	defer bq.ErrDefer()

	err = func() (err error) {
		u, e := Mkdir(ctx, m.Target, 0o777, true)
		if err = e; err != nil {
			return
		}
		bq.Put(u)

		if err = syscall.Mount(m.Source, m.Target, m.Fstype, m.Flags, m.Data); err != nil {
			return
		}
		bq.Put(mkUnmount(ctx, m.Target, "unmount"))
		return
	}()

	err = checkErr(ctx, err, m.Optional, "mount",
		slog.String("target", m.Target),
		slog.String("source", m.Source),
		slog.Bool("optional", m.Optional))
	return
}

func mkUnmount(ctx context.Context, target, act string) Undo {
	return func() {
		if err := syscall.Unmount(target, syscall.MNT_DETACH|syscall.MNT_FORCE); err != nil {
			slog.DebugContext(ctx, act+" fail", "target", target, "err", Errno(err).Error())
		}
	}
}
