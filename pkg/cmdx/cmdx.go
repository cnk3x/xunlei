package cmdx

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/log"
)

type Cmd struct {
	*cmd
	preStart  func(*Cmd) error
	onStarted func(*Cmd) error
	onExit    func(*Cmd) error
}

type cmd = exec.Cmd

func Exec(ctx context.Context, name string, options ...Option) (err error) {
	defer log.LogDone(ctx, slog.LevelDebug, filepath.Base(name), &err).Defer()
	ctx, cancel := context.WithCancelCause(ctx)

	c := &Cmd{cmd: exec.CommandContext(ctx, name)}
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	c.Cancel = func() error { return syscall.Kill(-c.Process.Pid, syscall.SIGKILL) }

	undo, err := apply(Options(options...), c)
	if err != nil {
		slog.ErrorContext(ctx, "apply options fail", "err", err)
		err = fmt.Errorf("apply options fail: %w", err)
		cancel(err)
		return
	}
	defer undo()

	if c.preStart != nil {
		if err = c.preStart(c); err != nil {
			slog.ErrorContext(ctx, "pre start fail", "err", err)
			err = fmt.Errorf("pre start fail")
			cancel(err)
			return
		}
	}

	err = func() (err error) {
		if err = c.Start(); err != nil {
			slog.ErrorContext(ctx, "start fail", "command", c.String(), "dir", c.Dir, "err", err)
			err = fmt.Errorf("start fail: %w", err)
			cancel(err)
			return
		}
		slog.DebugContext(ctx, "started", "command", c.String(), "dir", c.Dir, "pid", c.Process.Pid)

		if c.onStarted != nil {
			if err = c.onStarted(c); err != nil {
				slog.ErrorContext(ctx, "post_start fail", "err", err)
				err = fmt.Errorf("post_start fail: %w", err)
				cancel(err)
			}
		}

		if e := c.Wait(); err == nil && e != nil {
			err = e
		}
		return
	}()

	if c.onExit != nil {
		if e := c.onExit(c); e != nil {
			slog.ErrorContext(ctx, "post_exit fail", "err", e)
			if err == nil {
				err = e
			}
		}
	}

	if err != nil {
		slog.ErrorContext(ctx, "done", "err", err)
	} else {
		slog.DebugContext(ctx, "done")
	}

	cancel(err)
	return
}
