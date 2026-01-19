package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cnk3x/xunlei"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/cnk3x/xunlei/pkg/vms"
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
	slog.InfoContext(ctx, fmt.Sprintf("chroot: %s", cfg.Root))
	slog.InfoContext(ctx, fmt.Sprintf("spk_url: %s", cfg.SpkUrl))
	slog.InfoContext(ctx, fmt.Sprintf("force_download: %t", cfg.ForceDownload))

	err := vms.Exec(
		log.Prefix(ctx, "vms"),

		vms.Before(xunlei.Before(cfg)),
		vms.Run(xunlei.Run(cfg)),
		vms.Debug(cfg.Debug),
		vms.User(cfg.Uid, cfg.Gid),
		vms.Root(cfg.Root),
		vms.Binds("/lib", "/bin", "/etc/ssl"),
		vms.Links("/etc/timezone", "/etc/localtime", "/etc/resolv.conf"),
		vms.Links("/etc/passwd", "/etc/group", "/etc/shadow"),
		vms.Basic,
	)

	if err != nil {
		slog.ErrorContext(ctx, "exit", "err", err)
	} else {
		slog.InfoContext(ctx, "exit")
	}
}
