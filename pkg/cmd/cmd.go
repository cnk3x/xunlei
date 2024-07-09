package cmd

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"os/exec"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/log"
	"golang.org/x/sync/errgroup"
)

type RunOption struct {
	Dir string
	Env EnvSet
	Tag string
	Log func(ctx context.Context, s string)
}

func Run(ctx context.Context, name string, args []string, opts RunOption) (err error) {
	xlc := exec.CommandContext(ctx, name, args...)
	SetupProcAttr(xlc, syscall.SIGINT, 0, 0)

	xlc.Dir = opts.Dir
	xlc.Env = opts.Env

	ctx = log.Prefix(ctx, opts.Tag)

	var epr, spr io.ReadCloser
	if opts.Log != nil {
		xlc.Stderr = nil
		epr, _ = xlc.StderrPipe()

		xlc.Stdout = nil
		spr, _ = xlc.StdoutPipe()
	}

	slog.InfoContext(ctx, "start", "cmd", xlc.String())
	for _, it := range xlc.Env {
		slog.DebugContext(ctx, "env - "+it)
	}

	if err = xlc.Start(); err != nil {
		return
	}

	slog.InfoContext(ctx, "started", "pid", xlc.Process.Pid)

	if opts.Log != nil {
		readLog := func(ctx context.Context, r io.Reader) (err error) {
			for br := bufio.NewScanner(r); br.Scan(); {
				opts.Log(ctx, br.Text())
			}
			return
		}
		g, c := errgroup.WithContext(ctx)
		g.Go(func() error { return readLog(c, spr) })
		g.Go(func() error { return readLog(c, epr) })
		if e := g.Wait(); e != nil {
			slog.WarnContext(ctx, "console done", "err", e)
		} else {
			slog.InfoContext(ctx, "console done")
		}
	}

	if err = xlc.Wait(); err != nil {
		if IsSignaled(err) {
			err = nil
		}
	}

	slog.ErrorContext(ctx, "exited", "err", err)
	return
}
