package sys

import (
	"context"
	"encoding/json"
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

// 完整的绑定
func Mount(ctx context.Context, m MountOptions) (undo Undo, err error) {
	defer func() {
		if err != nil {
			slog.DebugContext(ctx, string(utils.Eon(json.Marshal(m))))
		}
	}()

	u := utils.MakeUndoPool(&undo, &err)
	defer u.ErrDefer()

	var dirUndo Undo
	if dirUndo, err = Mkdir(ctx, m.Target, 0777); err == nil {
		u.Put(dirUndo)
		if err = syscall.Mount(m.Source, m.Target, m.Fstype, m.Flags, m.Data); err == nil {
			u.Put(mkUnmount(ctx, m.Target, "unmount"))
		}
	}

	err = logIt(ctx, err, m.Optional, "mount",
		slog.String("target", m.Target),
		slog.String("source", m.Source),
		slog.Bool("optional", m.Optional))
	return
}

func mkUnmount(ctx context.Context, target, act string) Undo {
	return func() {
		err := syscall.Unmount(target, syscall.MNT_DETACH|syscall.MNT_FORCE)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelWarn, act, slog.String("target", target), slog.String("err", err.Error()))
			return
		}
		slog.LogAttrs(ctx, slog.LevelDebug, act, slog.String("target", target))
	}
}
