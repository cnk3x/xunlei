package sys

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"syscall"
)

// RunAs 临时切换到指定UID/GID身份执行自定义任务，妥善处理身份切换失败场景
//
//	uid 目标普通用户ID
//	gid 目标普通用户组ID
//	run 自定义任务函数
func RunAs(ctx context.Context, uid, gid int, run func() error) error {
	oEUid := syscall.Geteuid()
	oEGid := syscall.Getegid()

	if uid == oEUid && gid == oEGid {
		return run()
	}

	oUid := syscall.Getuid()
	if oUid != 0 {
		return errors.New("runas: only the root process (UID=0) supports.")
	}

	defer func() {
		if err := syscall.Seteuid(oEUid); err != nil {
			slog.DebugContext(ctx, "runas: restore uid fail", "uid", oEUid, "err", err)
		}

		if err := syscall.Setegid(oEGid); err != nil {
			slog.DebugContext(ctx, "runas: restore gid fail", "gid", oEGid, "err", err)
		}
	}()

	if err := syscall.Setegid(gid); err != nil {
		return fmt.Errorf("runas: change uid(%d) fail：%w", gid, err)
	}
	slog.InfoContext(ctx, "runas: changed uid(%d)", "gid", gid)

	if err := syscall.Seteuid(uid); err != nil {
		return fmt.Errorf("runas: change gid(%d) fail：%w", uid, err)
	}
	slog.InfoContext(ctx, "runas: changed gid(%d)", "uid", uid)

	return run()
}
