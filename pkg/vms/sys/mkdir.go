package sys

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/cnk3x/xunlei/pkg/utils"
)

func Mkdirs(ctx context.Context, dirs []string, perm fs.FileMode) (undo Undo, err error) {
	return doMulti(ctx, dirs,
		func(ctx context.Context, dir string) (Undo, error) {
			return Mkdir(ctx, dir, perm)
		},
	)
}

func MkdirsTask(ctx context.Context, dirs []string, perm fs.FileMode) func() (undo Undo, err error) {
	return func() (undo Undo, err error) { return Mkdirs(ctx, dirs, perm) }
}

func MkdirTask(ctx context.Context, dir string, perm fs.FileMode) func() (undo Undo, err error) {
	return func() (undo Undo, err error) { return Mkdir(ctx, dir, perm) }
}

// 脱了裤子放个屁，为了能够方便回滚
//   - 同 [os.MkdirAll] 创建路径中所有不存在的文件夹
//   - 不同的是，记录了创建的文件夹，并允许回滚
//   - 创建失败时，会自动回滚已创建过的上级文件夹
func Mkdir(ctx context.Context, dir string, perm fs.FileMode, logDisabled ...bool) (undo Undo, err error) {
	bq := utils.BackQueue(&undo, &err)
	defer bq.ErrDefer()

	if dir, err = filepath.Abs(dir); err != nil {
		return
	}

	logIt := slog.DebugContext
	if len(logDisabled) > 0 && logDisabled[0] {
		logIt = func(ctx context.Context, msg string, args ...any) {}
	}

	for _, cur := range allNotExists(dir) {
		if err = os.Mkdir(cur, perm); err != nil {
			slog.DebugContext(ctx, "mkdir fail", "dir", cur, "err", Errno(err))
			return
		}
		logIt(ctx, "mkdir", "perm", Perm2s(perm), "dir", dir)
		bq.Put(newRm(ctx, dir, "rmdir"))
	}

	if err != nil {
		slog.DebugContext(ctx, "mkdir fail", "dir", dir, "err", Errno(err))
	}

	return
}

// notice: dir must abs
func allNotExists(dir string) (notExistDirs []string) {
	for p, i := dir, 0; i < 100; i++ {
		if _, e := os.Stat(p); e == nil {
			break
		}
		notExistDirs = append(notExistDirs, p)
		c := filepath.Dir(p)
		if c == p {
			break
		}
		p = c
	}
	slices.Reverse(notExistDirs)
	return
}
