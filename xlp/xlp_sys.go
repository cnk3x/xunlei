//go:build linux && amd64

package xlp

import (
	"os"
	"os/exec"
	"syscall"
)

func setupProcAttr(c *exec.Cmd) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Pdeathsig = syscall.SIGKILL
}

func checkEnv() (err error) {
	if stat, _ := os.Stat("/.dockerenv"); stat != nil && stat.Mode().IsRegular() {
		return os.Remove("/.dockerenv")
	}
	return
}
