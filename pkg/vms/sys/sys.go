package sys

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/utils"
)

type Undo = func()

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

func logIt(ctx context.Context, err error, optional bool, name string, attrs ...slog.Attr) error {
	if err != nil {
		switch {
		case errors.Is(err, syscall.ENOTEMPTY):
			err = syscall.ENOTEMPTY
		case os.IsNotExist(err):
			err = os.ErrNotExist
		case os.IsExist(err):
			err = os.ErrExist
		}
		attrs = append(attrs, slog.String("err", err.Error()))
		if optional {
			slog.LogAttrs(ctx, slog.LevelDebug, name+" skip", attrs...)
		} else {
			slog.LogAttrs(ctx, slog.LevelWarn, name+" fail", attrs...)
		}
	} else {
		slog.LogAttrs(ctx, slog.LevelDebug, name, attrs...)
	}
	return utils.Iif(optional, nil, err)
}
