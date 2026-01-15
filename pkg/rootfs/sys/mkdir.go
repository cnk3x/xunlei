package sys

import (
	"cmp"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

	var undos []Undo
	undo = Undos(&undos)
	defer ExecUndo(undo, &err)

	vol := cmp.Or(filepath.VolumeName(dir), "/")
	items := strings.FieldsFunc(strings.TrimPrefix(dir, vol), func(r rune) bool { return r == '/' })

	var ok bool
	for i := range items {
		ok, err = mkdir(filepath.Join(vol, filepath.Join(items[:i+1]...)), perm)
	}

	switch {
	case err != nil:
		slog.WarnContext(ctx, "mkdir fail", "dir", dir, "err", err)
		return
	case ok:
		undos = append(undos, newRm(ctx, dir, "rmdir"))
		slog.DebugContext(ctx, "mkdir done", "dir", dir)
		// default:
		// 	slog.DebugContext(ctx, "mkdir skip", "dir", dir)
	}

	return
}

func newRm(ctx context.Context, target, act string) func() {
	return func() {
		err := os.Remove(target)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelWarn, act, slog.String("target", target), slog.String("err", err.Error()))
			return
		}
		slog.LogAttrs(ctx, slog.LevelDebug, act, slog.String("target", target))
	}
}

func mkdir(dir string, perm fs.FileMode) (ok bool, err error) {
	stat, e := os.Stat(dir)
	//存在
	if err = e; err == nil {
		if stat.IsDir() { //目录存在, 跳过
			return
		}

		if stat.Mode()&os.ModeSymlink != 0 { //软链接, 删除
			err = os.Remove(dir)
		} else { //文件存在, 报错
			err = fmt.Errorf("%w: %s", os.ErrExist, dir)
		}

		if err != nil {
			return
		}
	}

	//非文件不存在, 报错
	if !os.IsNotExist(err) {
		return
	}

	//创建目录, 上级目录存在
	err = os.Mkdir(dir, perm)
	ok = err == nil
	return
}
