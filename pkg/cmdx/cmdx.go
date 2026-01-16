package cmdx

import (
	"context"
	"os"
	"os/exec"
	"syscall"
)

var currentUid = os.Getuid()

func IsRoot() bool { return currentUid == 0 }

func Exec(ctx context.Context, name string, options ...CmdOption) error {
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
	return c.Run()
}
