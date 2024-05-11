package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/cnk3x/xunlei/xlp"
)

func main() {
	var daemon = xlp.New().WithFlag(flag.CommandLine)
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	daemon.Run(ctx)
}
