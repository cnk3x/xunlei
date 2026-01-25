package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cnk3x/flags"
	"github.com/cnk3x/xunlei"
	"github.com/cnk3x/xunlei/pkg/log"
)

var BuildTime string

func main() {
	cfg := &xunlei.Config{}
	_ = xunlei.ConfigCheck(cfg)

	fSet := flags.NewSet(
		flags.Version(xunlei.Version),
		flags.BuildTime(BuildTime),
		flags.Description("迅雷远程下载服务(非官方)"),
	)
	fSet.Struct(cfg)
	if !fSet.Parse() {
		os.Exit(1)
	}

	if err := xunlei.ConfigCheck(cfg); err != nil {
		slog.Error("exit", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.ForDefault(cfg.Log)
	ctx = log.Prefix(ctx, "main")

	xunlei.Banner(func(s string) { slog.Info(s) })
	slog.InfoContext(ctx, "VERSION: "+fSet.Version())
	slog.InfoContext(ctx, "BUILD_TIME: "+fSet.BuildTime().In(time.Local).Format(time.RFC3339))
	xunlei.ConfigPrint(cfg, func(s string) { slog.InfoContext(ctx, "XL_"+s) })

	if err := xunlei.Run(ctx, cfg); err != nil {
		slog.ErrorContext(ctx, "exit", "err", err)
	} else {
		slog.InfoContext(ctx, "exit")
	}
}
