package cmdx

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type CmdOption func(*exec.Cmd) (func(), error)
type ProcAttrOption func(*syscall.SysProcAttr) error

func neOpt(options ...func(c *exec.Cmd)) CmdOption {
	return func(c *exec.Cmd) (func(), error) {
		for _, option := range options {
			option(c)
		}
		return nil, nil
	}
}

func ProcAttr(sets ...ProcAttrOption) CmdOption {
	return func(c *exec.Cmd) (func(), error) {
		if c.SysProcAttr == nil {
			c.SysProcAttr = &syscall.SysProcAttr{}
		}
		for _, set := range sets {
			if err := set(c.SysProcAttr); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
}

func Credential(uid, gid uint32) CmdOption {
	return ProcAttr(func(spa *syscall.SysProcAttr) error {
		if uid != 0 || gid != 0 {
			if !IsRoot() {
				return fmt.Errorf("%s: root required, current uid: %d", os.ErrPermission, currentUid)
			}
			spa.Credential = &syscall.Credential{Uid: uid, Gid: gid, NoSetGroups: true}
		}
		return nil
	})
}

func Args(args ...string) CmdOption {
	return neOpt(func(c *exec.Cmd) { c.Args = append(c.Args[:1], args...) })
}

func Dir(dir string) CmdOption   { return neOpt(func(c *exec.Cmd) { c.Dir = dir }) }
func Env(env []string) CmdOption { return neOpt(func(c *exec.Cmd) { c.Env = env }) }
func Std() CmdOption             { return neOpt(func(c *exec.Cmd) { c.Stderr, c.Stdout = os.Stderr, os.Stdout }) }

func LineRead(lineRead func(string)) CmdOption {
	return func(c *exec.Cmd) (func(), error) {
		w := LineWriter(lineRead)
		c.Stdout, c.Stderr = w, w
		return func() { w.Close() }, nil
	}
}
