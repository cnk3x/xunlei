package sys

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cnk3x/xunlei/pkg/fo"
	"github.com/cnk3x/xunlei/pkg/utils"
)

func CreateFile(ctx context.Context, fn string, process fo.Process, fopts ...fo.Option) (undo func(), err error) {
	bq := utils.BackQueue(&undo, &err)
	defer bq.ErrDefer()

	err = func() (err error) {
		d, e := Mkdir(ctx, filepath.Dir(fn), 0777, true)
		if err = e; err != nil {
			return
		}
		bq.Put(d)

		if err = fo.OpenWrite(fn, process, append(fopts, fo.FlagExcl(false))...); err != nil {
			return
		}
		bq.Put(newRm(ctx, fn, "rmfile"))

		slog.DebugContext(ctx, "create", "path", fn)
		return
	}()

	if err != nil && !os.IsExist(err) {
		slog.DebugContext(ctx, "create fail", "path", fn, "err", err.Error())
		return
	}

	err = nil
	return
}

func FileCreateTask(ctx context.Context, fn string, process fo.Process, fopts ...fo.Option) func() (undo func(), err error) {
	return func() (undo func(), err error) {
		return CreateFile(ctx, fn, process, fopts...)
	}
}
