package sys

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/utils"
)

type Undo = func()

// 给路径添加前缀路径
func AddRoot(root string, paths ...string) []string {
	if root != "" {
		for i, p := range paths {
			paths[i] = filepath.Join(root, p)
		}
	}
	return paths
}

func Perm2s(perm os.FileMode) string {
	// return fmt.Sprintf("%s(0%s)", perm.Perm().String(), strconv.FormatInt(int64(perm.Perm()), 8))
	return "0" + strconv.FormatInt(int64(perm.Perm()), 8)
}

func Errno(err error) error {
	if err != nil {
		var errno syscall.Errno
		if errors.As(err, &errno) {
			return errno
		}
	}
	return err
}

func doMulti[O any](ctx context.Context, items []O, itemFn func(context.Context, O) (Undo, error)) (undo Undo, err error) {
	bq := utils.BackQueue(&undo, &err)
	defer bq.ErrDefer()

	for _, item := range items {
		u, e := itemFn(ctx, item)
		if err = e; e != nil {
			return
		}
		bq.Put(u)
	}
	return
}

func checkErr(ctx context.Context, err error, optional bool, act string, attrs ...slog.Attr) error {
	if err == nil {
		slog.LogAttrs(ctx, slog.LevelDebug, act, attrs...)
		return nil
	}

	attrs = append(attrs, slog.Bool("optional", optional), slog.String("err", Errno(err).Error()))

	if !optional {
		slog.LogAttrs(ctx, slog.LevelDebug, act+" fail", attrs...)
		return err
	}

	// if os.IsExist(err) || os.IsNotExist(err) {
	// 	return nil
	// }

	slog.LogAttrs(ctx, slog.LevelDebug, act+" skip", attrs...)
	return nil
}

func newRm(ctx context.Context, target, act string) Undo {
	return func() {
		if err := os.Remove(target); act != "" {
			if err != nil {
				slog.DebugContext(ctx, act+" fail", "err", Errno(err).Error())
			}
		}
	}
}
