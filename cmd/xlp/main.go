package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/cnk3x/xunlei"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/rootfs"
	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/cnk3x/xunlei/spk"
)

func main() {
	var cfg xunlei.Config
	if err := xunlei.ConfigBind(&cfg); err != nil {
		slog.Error("exit", "err", err)
		os.Exit(1)
	}

	log.ForDefault(utils.Iif(cfg.Debug, "debug", "info"), false)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ctx = log.Prefix(ctx, "main")
	slog.InfoContext(ctx, `_  _ _  _ _  _ _    ____  _`)
	slog.InfoContext(ctx, ` \/  |  | |\ | |    |___  |`)
	slog.InfoContext(ctx, `_/\_ |__| | \| |___ |___  |`)
	slog.InfoContext(ctx, fmt.Sprintf(`%-14s %s`, "daemon version:", xunlei.Version))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %t", "debug:", cfg.Debug))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %d", "port:", cfg.Port))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %s", "ip:", cfg.Ip))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %s", "dashboard username:", cfg.DashboardUsername))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %s", "dashboard password:", cfg.DashboardPassword))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %s", "dir download:", cfg.DirDownload))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %s", "dir data:", cfg.DirData))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %d", "uid:", cfg.Uid))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %d", "gid:", cfg.Gid))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %t", "prevent update:", cfg.PreventUpdate))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %s", "chroot:", cfg.Chroot))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %s", "spk_url:", cfg.SpkUrl))
	slog.InfoContext(ctx, fmt.Sprintf("%-14s %t", "force_download:", cfg.ForceDownload))

	if err := spk.Download(ctx, cfg.SpkUrl, filepath.Join(cfg.Chroot, xunlei.SYNOPKG_PKGDEST), cfg.ForceDownload); err != nil {
		return
	}

	if spkVer := utils.Cat(filepath.Join(cfg.Chroot, xunlei.PAN_XUNLEI_VER)); spkVer != "" {
		slog.InfoContext(ctx, fmt.Sprintf(`%-14s %s`, "spk version:", spkVer))
	}

	if cliVer := utils.Cat(filepath.Join(cfg.DirData, ".drive", "bin", "version")); cliVer != "" {
		slog.InfoContext(ctx, fmt.Sprintf(`%-14s %s`, "xunlei version:", cliVer))
	}

	err := rootfs.Run(ctx,
		cfg.Chroot, xunlei.NewRun(cfg),
		rootfs.Basic,
		rootfs.MountBindRoot("/lib", cfg.Chroot),
		rootfs.MountBindRoot("/lib64", cfg.Chroot, rootfs.Optional()),
		rootfs.MountBindRoot("/etc/ssl", cfg.Chroot, rootfs.Optional()),
		rootfs.LinkRoot("/etc/timezone", cfg.Chroot, 0666, true),
		rootfs.LinkRoot("/etc/resolv.conf", cfg.Chroot, 0666, true),
	)

	if err != nil {
		slog.ErrorContext(ctx, "exit", "err", err)
		os.Exit(1)
	}

	slog.InfoContext(ctx, "exit")
}
