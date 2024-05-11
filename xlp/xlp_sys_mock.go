//go:build !linux || !amd64

package xlp

import (
	"fmt"
	"os/exec"
	"runtime"
)

func setupProcAttr(c *exec.Cmd) {}

func checkEnv() (err error) {
	return fmt.Errorf("unsupported platform: %s %s", runtime.GOOS, runtime.GOARCH)
}
