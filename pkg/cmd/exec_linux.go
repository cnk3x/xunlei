//go:build linux
// +build linux

package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os/exec"
	"os/user"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/lod"
)

func SetupProcAttr(c *exec.Cmd, pdeathsig syscall.Signal, uid, gid uint32) {
	Setpgid(c)
	SetPdeathsig(c, pdeathsig)
	SetCredential(c, uid, gid)
}

func IsSignaled(err error) bool {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if status, ok := ee.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			return status.Signaled()
		}
	}
	return false
}

func SetCredential(c *exec.Cmd, uid, gid uint32) {
	if uid != 0 && gid != 0 { //必须有root权限才能设置
		if c.SysProcAttr == nil {
			c.SysProcAttr = &syscall.SysProcAttr{}
		}
		c.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, NoSetGroups: true}
	}
}

func SetNewNS(ctx context.Context, c *exec.Cmd) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Cloneflags |= syscall.CLONE_NEWNS
	slog.DebugContext(ctx, "SetNewNS")
}

func SetPdeathsig(c *exec.Cmd, sig syscall.Signal) {
	if sig != 0 {
		if c.SysProcAttr == nil {
			c.SysProcAttr = &syscall.SysProcAttr{}
		}
		c.SysProcAttr.Pdeathsig = sig
		//slog.DebugContext(ctx, "SetPdeathsig", "signal", sig)
	}
}

func Setpgid(c *exec.Cmd) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Setpgid = true
	//slog.DebugContext(ctx, "Setpgid")
}

func isStrZero(s string) bool { return len(s) == 0 || s == "0" }

func User(userNameOrId, groupNameOrId string) (uid, gid uint32, err error) {
	if isStrZero(userNameOrId) && isStrZero(groupNameOrId) {
		return
	}

	u, ue := ParseUserId(userNameOrId)
	if ue != nil {
		err = ue
		return
	}

	if g, ge := ParseGroupId(groupNameOrId); g != nil && ge == nil {
		u.Gid = g.Gid
	}

	uid = lod.May(lod.ParseUint32(u.Uid))
	gid = lod.May(lod.ParseUint32(u.Gid))
	return
}

func ParseUserId(nameOrId string) (u *user.User, err error) {
	if nameOrId == "" {
		nameOrId = "0"
	}

	if u, err = user.Current(); err == nil && u != nil && (u.Username == nameOrId || u.Uid == nameOrId) {
		return
	}

	if u, err = user.Lookup(nameOrId); err == nil {
		return
	}

	if u, err = user.LookupId(nameOrId); err == nil {
		return
	}

	uid, e := lod.ParseUint32(nameOrId)
	if e != nil {
		return nil, e
	}

	u = &user.User{Uid: lod.FormatInt(uid)}
	return
}

func ParseGroupId(nameOrId string) (g *user.Group, err error) {
	if nameOrId == "" {
		nameOrId = "0"
	}

	if g, err = user.LookupGroup(nameOrId); err == nil {
		return
	}

	if g, err = user.LookupGroupId(nameOrId); err == nil {
		return
	}

	gid, e := lod.ParseUint32(nameOrId)
	if e != nil {
		err = e
		return
	}

	g = &user.Group{Gid: lod.FormatInt(gid)}
	return
}
