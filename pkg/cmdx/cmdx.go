package cmdx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"syscall"
)

type Cmd struct {
	*cmd
	ctx       context.Context
	preStart  func(*Cmd) error
	onStarted func(*Cmd) error
	onExit    func(*Cmd) error
}

type cmd = exec.Cmd

func (c *Cmd) Context() context.Context { return c.ctx }

func Run(ctx context.Context, name string, options ...Option) (err error) {
	c := &Cmd{ctx: ctx, cmd: exec.CommandContext(ctx, name)}
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.Cancel = func() error { return syscall.Kill(-c.Process.Pid, syscall.SIGKILL) }

	defer func() {
		if c.onExit != nil {
			if e := c.onExit(c); err == nil && e != nil {
				err = e
			}
		}
	}()

	clean, err := Options(options...)(c)
	if err != nil {
		err = fmt.Errorf("cmdx options: %w", err)
		return
	}
	defer clean()

	if c.preStart != nil {
		if err = c.preStart(c); err != nil {
			err = fmt.Errorf("cmdx pre start: %w", err)
			return
		}
	}

	err = func() (err error) {
		if err = c.Start(); err != nil {
			slog.DebugContext(ctx, "start", "command", c.String(), "dir", c.Dir)
			return fmt.Errorf("cmdx start: %w", err)
		}
		slog.DebugContext(ctx, "cmdx started", "command", c.String(), "dir", c.Dir, "pid", c.Process.Pid)

		if c.onStarted != nil {
			if err = c.onStarted(c); err != nil {
				return fmt.Errorf("cmdx started: %w", err)
			}
		}

		if err = c.Wait(); err != nil {
			if signal := GetSignal(err); signal != -1 {
				slog.DebugContext(ctx, "cmdx signaled: %s(%d)", signal.String(), signal)
				return
			}
			return fmt.Errorf("cmdx wait: %w", err)
		}

		return
	}()

	return
}

// GetSignal 判断 cmd.Wait() 返回的错误是否为信号中断，返回-1代表不是。
func GetSignal(err error) syscall.Signal {
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		return -1
	}
	ws, ok := ee.Sys().(syscall.WaitStatus)
	if !ok || !ws.Signaled() {
		return -1
	}
	return ws.Signal()
}
