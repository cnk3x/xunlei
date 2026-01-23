package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cnk3x/flags"
	"github.com/cnk3x/xunlei"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/vms"
)

var BuildTime string

func main() {
	cfg := &xunlei.Config{}

	fSet := flags.NewSet(flags.Version(xunlei.Version), flags.BuildTime(BuildTime))
	fSet.Struct(cfg)
	if !fSet.Parse() {
		os.Exit(1)
	}

	if err := xunlei.ConfigCheck(cfg); err != nil {
		slog.Error("exit", "err", err)
		os.Exit(1)
	}

	xunlei.Banner(func(s string) { slog.Info(s) })
	xunlei.ConfigPrint(cfg, func(s string) { slog.Info("XL_" + s) })

	if vms.RootRequired(cfg.Root) {
		if uid := os.Getuid(); uid != 0 {
			slog.Error("exit", "err", "must run as root", "uid", uid)
			os.Exit(1)
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ctx = log.Prefix(ctx, "main")

	slog.InfoContext(ctx, "VERSION: "+fSet.Version())
	slog.InfoContext(ctx, "BUILD_TIME: "+fSet.BuildTime().In(time.Local).Format(time.RFC3339))
	xunlei.ConfigPrint(cfg, func(s string) { slog.InfoContext(ctx, s) })

	err := vms.Execute(
		log.Prefix(ctx, "vms"),
		vms.Wait(cfg.Debug),
		vms.User(cfg.Uid, cfg.Gid),
		vms.Root(cfg.Root),
		vms.Binds("/lib", "/bin", "/etc/ssl"),
		vms.Links("/etc/timezone", "/etc/localtime", "/etc/resolv.conf"),
		vms.Links("/etc/passwd", "/etc/group", "/etc/shadow"),
		vms.Symlink("lib", filepath.Join(cfg.Root, "lib64")),
		vms.Basic,

		vms.Before(xunlei.Before(cfg)),
		vms.Run(xunlei.Run(cfg)),
	)

	if err != nil {
		slog.ErrorContext(ctx, "exit", "err", err)
	} else {
		slog.InfoContext(ctx, "exit")
	}
}
