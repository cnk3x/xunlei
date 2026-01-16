package vms

import (
	"errors"
	"os"
	"strconv"
	"syscall"
	"testing"
)

func TestLink(t *testing.T) {
	if err := os.Link(`/usr/bin`, `/mnt/d/2`); err != nil {
		t.Logf("[%T]: %v", err, err)
		var errno syscall.Errno
		if errors.As(err, &errno) {
			t.Logf("errno(0x%s)", strconv.FormatInt(int64(errno), 16))
		}
		//ENOENT(0x2): 源不存在, EXDEV(0x12): 跨设备连接, EEXIST(0x11): 目标存在, EPERM(0x1):操作不被允许
		t.Logf("err.is(exdev): %t", errors.Is(err, syscall.EXDEV))
	}
}

func TestErrDefer(t *testing.T) {
	var err error
	defer func(err *error) { t.Logf("defer 2: %v", *err) }(&err)
	defer t.Logf("defer 1: %v", err)
	err = errors.New("test")
}
