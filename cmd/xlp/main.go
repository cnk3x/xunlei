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
	slog.InfoContext(ctx, fmt.Sprintf(`daemon version: %s`, xunlei.Version))
	slog.InfoContext(ctx, fmt.Sprintf("debug: %t", cfg.Debug))
	slog.InfoContext(ctx, fmt.Sprintf("port: %d", cfg.Port))
	slog.InfoContext(ctx, fmt.Sprintf("ip: %s", cfg.Ip))
	slog.InfoContext(ctx, fmt.Sprintf("dashboard username: %s", cfg.DashboardUsername))
	slog.InfoContext(ctx, fmt.Sprintf("dashboard password: %s", cfg.DashboardPassword))
	slog.InfoContext(ctx, fmt.Sprintf("dir download: %s", cfg.DirDownload))
	slog.InfoContext(ctx, fmt.Sprintf("dir data: %s", cfg.DirData))
	slog.InfoContext(ctx, fmt.Sprintf("uid: %d", cfg.Uid))
	slog.InfoContext(ctx, fmt.Sprintf("gid: %d", cfg.Gid))
	slog.InfoContext(ctx, fmt.Sprintf("prevent update: %t", cfg.PreventUpdate))
	slog.InfoContext(ctx, fmt.Sprintf("chroot: %s", cfg.Chroot))
	slog.InfoContext(ctx, fmt.Sprintf("spk_url: %s", cfg.SpkUrl))
	slog.InfoContext(ctx, fmt.Sprintf("force_download: %t", cfg.ForceDownload))

	if err := spk.Download(ctx, cfg.SpkUrl, filepath.Join(cfg.Chroot, xunlei.SYNOPKG_PKGDEST), cfg.ForceDownload); err != nil {
		return
	}

	if spkVer := utils.Cat(filepath.Join(cfg.Chroot, xunlei.PAN_XUNLEI_VER)); spkVer != "" {
		slog.InfoContext(ctx, fmt.Sprintf(`spk version: %s`, spkVer))
	}

	if cliVer := utils.Cat(filepath.Join(cfg.DirData, ".drive", "bin", "version")); cliVer != "" {
		slog.InfoContext(ctx, fmt.Sprintf(`xunlei version: %s`, cliVer))
	}

	err := rootfs.Run(ctx,
		cfg.Chroot,
		xunlei.NewRun(cfg),
		rootfs.Before(xunlei.BeforeChroot(cfg)),
		rootfs.Basic,
		rootfs.MountBindRoot("/lib", cfg.Chroot),
		rootfs.MountBindRoot("/usr/lib", cfg.Chroot),
		rootfs.MountBindRoot("/lib64", cfg.Chroot, rootfs.Optional()),
		rootfs.MountBindRoot("/usr/lib64", cfg.Chroot, rootfs.Optional()),
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
