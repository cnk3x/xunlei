package sys

import (
	"context"
	"log/slog"

	"github.com/cnk3x/xunlei/pkg/utils"
)

type Undo = func()

func doMulti[O any](ctx context.Context, items []O, itemFn func(context.Context, O) (Undo, error)) (undo Undo, err error) {
	p := utils.MakeUndoPool(&undo, &err)
	defer p.ErrDefer()

	for _, item := range items {
		u, e := itemFn(ctx, item)
		if err = e; e != nil {
			return
		}
		p.Put(u)
	}
	return
}

func logIt(ctx context.Context, err error, optional bool, name string, attrs ...slog.Attr) error {
	if err != nil {
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
