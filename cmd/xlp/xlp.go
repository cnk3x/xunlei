package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/cnk3x/xunlei/xlp"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: "01/02 15:04:05"})

	var daemon = xlp.New().WithFlag(flag.CommandLine)
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	daemon.Run(ctx)
}
