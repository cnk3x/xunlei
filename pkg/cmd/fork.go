package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
)

/* fork */

const keyTag = "GO_CMD_FORK_TAG"

func ForkTag() string {
	return os.Getenv(keyTag)
}

func IsForkTag(tag string) bool {
	return os.Getenv(keyTag) == tag
}

type ForkOptions struct {
	Args []string
	Env  EnvSet
	Uid  uint32
	Gid  uint32
	Tag  string
}

func Fork(ctx context.Context, fOpts ForkOptions) (err error) {
	cmd := exec.CommandContext(ctx, os.Args[0], fOpts.Args...)
	SetupProcAttr(cmd, syscall.SIGINT, fOpts.Uid, fOpts.Gid)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = fOpts.Env.Set(keyTag, fOpts.Tag)

	slog.DebugContext(ctx, fOpts.Tag+" run", "command", cmd.String(), "uid", fOpts.Uid, "gid", fOpts.Gid)
	if err = cmd.Start(); err != nil {
		return
	}

	slog.DebugContext(ctx, fOpts.Tag+" run", "pid", cmd.Process.Pid)

	if err = cmd.Wait(); err != nil {
		if IsSignaled(err) {
			return nil
		}
	}

	return
}

/* fork end */
