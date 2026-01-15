package sys

import (
	"context"
	"os"
	"syscall"
)

// chroot & run
func Chroot(ctx context.Context, newRoot string, run func(ctx context.Context) error) (err error) {
	wd, e := os.Getwd()
	if err = e; err != nil {
		return
	}

	if err = syscall.Chdir(newRoot); err != nil {
		return
	}
	defer syscall.Chdir(wd)

	var rfd int
	if rfd, err = syscall.Open("/", syscall.O_RDONLY, 0); err != nil {
		return
	}
	defer syscall.Close(rfd)

	if err = syscall.Chroot("."); err != nil {
		return
	}
	defer syscall.Chroot(".")
	defer syscall.Fchdir(rfd)

	if err = syscall.Chdir("/"); err != nil {
		return
	}

	err = run(ctx)
	return
}
