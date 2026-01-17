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
		return errors.New("RunAs: 仅root进程（UID=0）支持 可逆切换身份")
	}

	defer func() {
		if err := syscall.Seteuid(oEUid); err != nil {
			slog.DebugContext(ctx, "RunAs: 恢复有效UID失败", "uid", oEUid, "err", err)
		}

		if err := syscall.Setegid(oEGid); err != nil {
			slog.DebugContext(ctx, "RunAs: 恢复有效GID失败", "gid", oEGid, "err", err)
		}
	}()

	if err := syscall.Setegid(gid); err != nil {
		return fmt.Errorf("RunAs: 切换GID(%d)失败：%w", gid, err)
	}
	slog.DebugContext(ctx, "RunAs: 成功切换GID", "gid", gid)

	if err := syscall.Seteuid(uid); err != nil {
		return fmt.Errorf("RunAs: 切换UID(%d)失败：%w", uid, err)
	}
	slog.DebugContext(ctx, "RunAs: 成功切换UID", "uid", uid)

	return run()
}
