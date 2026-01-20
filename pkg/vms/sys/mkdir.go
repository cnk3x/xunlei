package sys

import (
	"cmp"
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/utils"
)

func Mkdirs(ctx context.Context, dirs []string, perm fs.FileMode) (undo Undo, err error) {
	return doMulti(ctx, dirs,
		func(ctx context.Context, dir string) (Undo, error) {
			return Mkdir(ctx, dir, perm)
		},
	)
}

// 脱了裤子放个屁，为了能够方便回滚
func Mkdir(ctx context.Context, dir string, perm fs.FileMode) (undo Undo, err error) {
	if dir, err = filepath.Abs(dir); err != nil {
		return
	}

	bq := utils.BackQueue(&undo, &err)
	defer bq.ErrDefer()

	vol := cmp.Or(filepath.VolumeName(dir), "/")
	items := strings.FieldsFunc(strings.TrimPrefix(dir, vol), func(r rune) bool { return r == '/' })

	var ok bool
	for i := range items {
		p := filepath.Join(vol, filepath.Join(items[:i+1]...))
		if ok, err = mkdir(p, perm); err != nil {
			slog.WarnContext(ctx, "mkdir fail", "dir", p, "err", err)
			break
		}

		if ok {
			bq.Put(newRm(ctx, p, utils.Iif(i == len(items)-1, "rmdir", "")))
		}
	}

	return
}

func newRm(ctx context.Context, target, act string) func() {
	return func() {
		if err := os.Remove(target); act != "" {
			logIt(ctx, err, true, act, slog.String("target", target))
		}
	}
}

func mkdir(dir string, perm fs.FileMode) (ok bool, err error) {
	if err = os.Mkdir(dir, perm); err == nil { //创建成功
		ok = true
		return
	}

	if errors.Is(err, syscall.EEXIST) {
		switch lstat, _ := os.Lstat(dir); {
		case lstat.IsDir(): //是文件夹，跳过
			err = nil
		case lstat.Mode()&os.ModeSymlink != 0: //是软链接，删除重建
			if err = os.Remove(dir); err == nil {
				ok, err = mkdir(dir, perm) //重新创建
			}
		}
	}
	return
}
