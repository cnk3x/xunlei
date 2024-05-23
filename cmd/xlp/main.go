package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"

	"github.com/cnk3x/xunlei/xlp"
	"github.com/lmittmann/tint"
)

var version = "unknown"

func main() {
	d := xlp.New().Version(version).BindFlag(flag.CommandLine, true)

	slog.SetDefault(
		slog.New(
			tint.NewHandler(
				os.Stderr,
				&tint.Options{
					TimeFormat: "01/02 15:04:05",
					Level:      xlp.Iif(d.IsDebug(), slog.LevelDebug, slog.LevelInfo),
					// AddSource:  d.IsDebug(),
				},
			),
		),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	d.Run(ctx)
}
