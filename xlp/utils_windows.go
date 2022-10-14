//go:build windows

package main

import (
	"os"
	"syscall"
)

func Chroot(path string) error {
	return syscall.ERROR_ACCESS_DENIED
}

func Chdir(dir string) error {
	return os.Chdir(dir)
}

func Mount(src, dst string) error {
	return syscall.ERROR_ACCESS_DENIED
}

func Unmount(path string) error {
	return syscall.ERROR_ACCESS_DENIED
}

func SetSysProc(attr *syscall.SysProcAttr) *syscall.SysProcAttr {
	attr.HideWindow = true
	return attr
}
