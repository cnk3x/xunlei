package cmdx

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"syscall"
)

type Cmd struct {
	exec.Cmd
	preStart  func(*Cmd) error
	onStarted func(*Cmd) error
	onExit    func(*Cmd) error
}

func Exec(ctx context.Context, name string, options ...Option) (err error) {
	ctx, cancel := context.WithCancelCause(ctx)

	c := &Cmd{Cmd: *(exec.CommandContext(ctx, name))}
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
			slog.ErrorContext(ctx, "pre_start fail", "err", err)
			err = fmt.Errorf("pre_start fail: %w", err)
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
