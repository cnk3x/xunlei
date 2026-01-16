package cmdx

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
)

var currentUid = os.Getuid()

func IsRoot() bool { return currentUid == 0 }

func Exec(ctx context.Context, name string, options ...CmdOption) (err error) {
	c := exec.CommandContext(ctx, name)
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.Cancel = func() error { return syscall.Kill(-c.Process.Pid, syscall.SIGKILL) }

	var closers []func()
	defer func() {
		for _, closer := range closers {
			closer()
		}
	}()
	for _, apply := range options {
		closer, err := apply(c)
		if err != nil {
			return err
		}
		if closer != nil {
			closers = append(closers, closer)
		}
	}

	if err = c.Start(); err != nil {
		slog.ErrorContext(ctx, "exec start fail", "command", c.String(), "dir", c.Dir, "err", err)
		return
	}
	slog.DebugContext(ctx, "exec started", "command", c.String(), "dir", c.Dir, "pid", c.Process.Pid)

	if err = c.Wait(); err != nil {
		slog.ErrorContext(ctx, "exec done", "err", err)
		return
	}

	slog.DebugContext(ctx, "exec done", "command", c.String(), "dir", c.Dir, "pid", c.Process.Pid)
	return
}
