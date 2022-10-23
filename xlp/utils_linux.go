//go:build unix

package main

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

func Chroot(path string) error {
	return syscall.Chroot(path)
}

func Chdir(dir string) error {
	return syscall.Chdir(dir)
}

func Mount(src, dst string) error {
	return syscall.Mount(src, dst, "auto", syscall.MS_BIND|syscall.MS_SLAVE|syscall.MS_REC, "")
}

func Unmount(path string) error {
	return syscall.Unmount(path, syscall.MNT_DETACH)
}

func SetSysProc(attr *syscall.SysProcAttr) *syscall.SysProcAttr {
	attr.Pdeathsig = syscall.SIGKILL
	return attr
}

func SetUser(attr *syscall.SysProcAttr) (uid, gid int) {
	euid := os.Getenv("UID")
	egid := os.Getenv("GID")
	uid, _ = strconv.Atoi(euid)
	gid, _ = strconv.Atoi(egid)
	fmt.Printf("uid=%d,gid=%d\n", uid, gid)
	attr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	return
}
