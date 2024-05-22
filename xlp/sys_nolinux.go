//go:build !linux

package xlp

import (
	"fmt"
	"os/exec"
	"runtime"
)

var ErrPlatformNotSupport = fmt.Errorf("platform %s-%s not supported", runtime.GOOS, runtime.GOARCH)

func setupProcAttr(*exec.Cmd, uint32, uint32)         {}
func setupWrapProcAttr(*exec.Cmd, string)             {}
func checkEnv() error                                 { return ErrPlatformNotSupport }
func lookupUg(string, string) (uint32, uint32, error) { return 0, 0, ErrPlatformNotSupport }

func sysMount(string, string) (err error) { return }
func sysUnmount(string) (err error)       { return }
func sysChroot(string) (err error)        { return }

var _ = setupWrapProcAttr
