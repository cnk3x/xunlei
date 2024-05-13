//go:build linux && (amd64 || arm64)
// +build linux
// +build amd64 arm64

package xlp

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/cnk3x/xunlei/embeds"
)

func setupProcAttr(c *exec.Cmd) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Pdeathsig = syscall.SIGKILL
}

func checkEnv() (err error) {
	if stat, _ := os.Stat("/.dockerenv"); stat != nil {
		return os.Remove("/.dockerenv")
	}
	return embeds.Extract(SYNOPKG_PKGDEST)
}
