package cmdx

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"syscall"
)

type Cmd struct {
	*cmd
	preStart  func(*Cmd) error
	onStarted func(*Cmd) error
	onExit    func(*Cmd) error
}

type cmd = exec.Cmd

func Run(ctx context.Context, name string, options ...Option) (err error) {
	c := &Cmd{cmd: exec.CommandContext(ctx, name)}
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	c.Cancel = func() error { return syscall.Kill(-c.Process.Pid, syscall.SIGKILL) }

	closes, err := Options(options...)(c)
	if err != nil {
		err = fmt.Errorf("cmdx options: %w", err)
		return
	}
	defer closes()

	if c.preStart != nil {
		if err = c.preStart(c); err != nil {
			err = fmt.Errorf("cmdx pre start: %w", err)
			return
		}
	}

	err = func() (err error) {
		if err = c.Start(); err != nil {
			slog.DebugContext(ctx, "start", "command", c.String(), "dir", c.Dir)
			err = fmt.Errorf("cmdx start: %w", err)
			return
		}
		slog.DebugContext(ctx, "started", "command", c.String(), "dir", c.Dir, "pid", c.Process.Pid)

		if c.onStarted != nil {
			if err = c.onStarted(c); err != nil {
				err = fmt.Errorf("cmdx started: %w", err)
			}
		}

		if e := c.Wait(); err == nil && e != nil {
			err = fmt.Errorf("cmdx wait: %w", e)
		}
		return
	}()

	if c.onExit != nil {
		if e := c.onExit(c); err == nil && e != nil {
			err = fmt.Errorf("cmdx exit: %w", e)
		}
	}
	return
}
