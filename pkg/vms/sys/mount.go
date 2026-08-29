package sys

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
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

	bq := utils.BackQueue(&undo, &err)
	defer bq.ErrDefer()

	var dirUndo Undo
	if dirUndo, err = Mkdir(ctx, m.Target, 0777); err == nil {
		bq.Put(dirUndo)
		err = syscall.Mount(m.Source, m.Target, m.Fstype, m.Flags, m.Data)
		if err == nil {
			bq.Put(mkUnmount(ctx, m.Target, "unmount"))
		} else if m.Fstype == "devtmpfs" && err == syscall.ENODEV {
			// Some minimal/custom kernels are built without CONFIG_DEVTMPFS.
			// In that case reuse the container's existing /dev recursively,
			// including child mounts such as /dev/pts and /dev/shm.
			devtmpfsErr := err
			err = syscall.Mount("/dev", m.Target, "", uintptr(syscall.MS_BIND|syscall.MS_REC), "")
			if err == nil {
				slog.WarnContext(ctx, "devtmpfs unavailable, fallback to rbind /dev",
					slog.String("target", m.Target),
					slog.String("err", devtmpfsErr.Error()))
				bq.Put(mkUnmount(ctx, m.Target, "unmount"))
			} else {
				err = fmt.Errorf("mount devtmpfs failed: %v; fallback rbind /dev failed: %w",
					devtmpfsErr, err)
			}
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
			if os.IsNotExist(err) {
				slog.LogAttrs(ctx, slog.LevelWarn, act, slog.String("target", target), slog.String("err", os.ErrNotExist.Error()))
			} else {
				slog.LogAttrs(ctx, slog.LevelWarn, act, slog.String("target", target), slog.String("err", err.Error()))
			}
			return
		}
		slog.LogAttrs(ctx, slog.LevelDebug, act, slog.String("target", target))
	}
}
