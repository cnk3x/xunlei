//go:build !windows && !darwin

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/cnk3x/xunlei"
	"github.com/cnk3x/xunlei/pkg/cmd"
	"github.com/cnk3x/xunlei/pkg/flags"
	"github.com/cnk3x/xunlei/pkg/lod"
	"github.com/cnk3x/xunlei/pkg/log"
)

var version = "unknown"

func main() {
	flags.Default.SetVersion(version)

	cfg := xunlei.ConfigDefault()
	flags.Struct(&cfg)
	flags.Parse()

	log.ForDefault(lod.Iif(cfg.Debug, "debug", "info"))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ctx = log.Prefix(ctx, lod.Select(cmd.ForkTag(), "main"))

	if err := xunlei.New(cfg, version).Run(ctx); err != nil {
		slog.ErrorContext(ctx, "exited!", "err", err.Error())
	} else {
		slog.InfoContext(ctx, "exited!")
	}
}
