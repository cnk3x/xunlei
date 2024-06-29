package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

func IsSignaled(err error) bool {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if status, ok := ee.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			return status.Signaled()
		}
	}
	return false
}

func SetCredential(ctx context.Context, c *exec.Cmd, uid, gid uint32) {
	idEq := func(id1 string, id2 uint32) bool { return id1 == strconv.FormatUint(uint64(id2), 10) }
	if u, _ := user.Current(); u != nil && idEq(u.Uid, uid) && idEq(u.Gid, gid) {
		if c.SysProcAttr == nil {
			c.SysProcAttr = &syscall.SysProcAttr{}
		}
		c.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, NoSetGroups: true}
		slog.DebugContext(ctx, "SetCredential", "uid", uid, "gid", gid)
	}
}

func SetNewNS(ctx context.Context, c *exec.Cmd) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Cloneflags |= syscall.CLONE_NEWNS
	slog.DebugContext(ctx, "SetNewNS")
}

func SetPdeathsig(ctx context.Context, c *exec.Cmd, sig syscall.Signal) {
	if sig != 0 {
		if c.SysProcAttr == nil {
			c.SysProcAttr = &syscall.SysProcAttr{}
		}
		c.SysProcAttr.Pdeathsig = sig
		slog.DebugContext(ctx, "SetPdeathsig", "signal", sig)
	}
}

func Setpgid(ctx context.Context, c *exec.Cmd) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Setpgid = true
	slog.DebugContext(ctx, "Setpgid")
}

func UserFind(userOrId, groupOrId string) (uid, gid uint32, err error) {
	var u *user.User
	var g *user.Group

	parseUint := func(s string) (uint32, error) {
		if out, err := strconv.ParseUint(s, 10, 32); err != nil {
			return 0, err
		} else {
			return uint32(out), nil
		}
	}

	if userOrId != "" {
		u, err = user.Lookup(userOrId)

		if err != nil {
			u, err = user.LookupId(userOrId)
		}

		if err != nil {
			uid, err = parseUint(userOrId)
		}

		if err != nil {
			return
		}
	}

	if groupOrId != "" {
		g, err = user.LookupGroup(groupOrId)

		if err != nil {
			g, err = user.LookupGroupId(groupOrId)
		}

		if err != nil {
			gid, err = parseUint(groupOrId)
		}

		if err != nil {
			return
		}
	}

	if u != nil {
		uid, _ = parseUint(u.Uid)
		gid, _ = parseUint(u.Gid)
	}

	if g != nil {
		gid, _ = parseUint(g.Gid)
	}

	return
}
