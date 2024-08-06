//go:build !linux
// +build !linux

package cmd

import (
	"context"
	"os/exec"
	"syscall"
)

func SetupProcAttr(c *exec.Cmd, pdeathsig syscall.Signal, uid, gid uint32) {}
func IsSignaled(err error) (ok bool)                                       { return }
func SetCredential(ctx context.Context, c *exec.Cmd, uid, gid uint32)      {}
func SetNewNS(ctx context.Context, c *exec.Cmd)                            {}
func SetPdeathsig(ctx context.Context, c *exec.Cmd, sig syscall.Signal)    {}
func Setpgid(ctx context.Context, c *exec.Cmd)                             {}
func UserFind(userOrId, groupOrId string) (uid, gid uint32, err error)     { return }
