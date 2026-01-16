package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/cnk3x/xunlei"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/cnk3x/xunlei/pkg/vms"
	"github.com/cnk3x/xunlei/spk"
)

func main() {
	var cfg xunlei.Config
	if err := xunlei.ConfigBind(&cfg); err != nil {
		slog.Error("exit", "err", err)
		os.Exit(1)
	}

	log.ForDefault(utils.Iif(cfg.Debug, "debug", "info"), false)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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
		slog.ErrorContext(ctx, "exit", "err", err)
		os.Exit(1)
	}

	if spkVer := utils.Cat(filepath.Join(cfg.Chroot, xunlei.PAN_XUNLEI_VER)); spkVer != "" {
		slog.InfoContext(ctx, fmt.Sprintf(`spk version: %s`, spkVer))
	}

	if cliVer := utils.Cat(filepath.Join(cfg.DirData, ".drive", "bin", "version")); cliVer != "" {
		slog.InfoContext(ctx, fmt.Sprintf(`xunlei version: %s`, cliVer))
	}

	err := vms.Run(log.Prefix(ctx, "boot"), cfg.Chroot, xunlei.NewRun(cfg),
		vms.Before(xunlei.NewBefore(cfg)), vms.Debug(cfg.Debug),
		vms.Basic,
		vms.Binds(cfg.Chroot, "/lib", "/bin", "/etc/ssl"),
		vms.Links(cfg.Chroot, "/etc/timezone", "/etc/localtime", "/etc/resolv.conf"),
		vms.Links(cfg.Chroot, "/etc/passwd", "/etc/group", "/etc/shadow"),
	)

	if err != nil {
		slog.ErrorContext(ctx, "exit", "err", err)
		os.Exit(1)
	}
	slog.InfoContext(ctx, "exit")
}
