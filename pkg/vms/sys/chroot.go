package sys

import (
	"cmp"
	"context"
	"log/slog"
	"os"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/utils"
)

// chroot & run
func Chroot(ctx context.Context, newRoot string, run func(ctx context.Context) error, debug ...bool) (err error) {
	err = func() (err error) {
		var wd string
		if wd, err = os.Getwd(); err != nil {
			return
		}
		defer log.LogDone(ctx, slog.LevelDebug, "chroot ", &err, "root", newRoot, "wd", wd)

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

		defer func() {
			err = utils.SeqExec(
				func() error { return syscall.Fchdir(rfd) },
				func() error { return syscall.Chroot(".") },
				func() error { return syscall.Chdir(wd) },
			)
			slog.Log(ctx, log.WarnDebug(err), "back to old root", "fd", rfd, "wd", wd, "err", err)
		}()

		if err = syscall.Chdir("/"); err != nil {
			return
		}

		slog.DebugContext(ctx, "run start")
		err = run(ctx)
		return
	}()

	if cmp.Or(debug...) {
		<-ctx.Done()
	}
	return
}
