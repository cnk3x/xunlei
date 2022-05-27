package main

import (
	"context"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if len(os.Args) > 1 && os.Args[1] == "syno" {
		syno(ctx)
	} else {
		xlp(ctx, getOpts())
	}
}
