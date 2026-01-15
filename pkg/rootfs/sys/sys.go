package sys

import (
	"context"
	"log/slog"
	"slices"

	"github.com/cnk3x/xunlei/pkg/utils"
)

type Undo = func()

func doMulti[O any](ctx context.Context, items []O, itemFn func(context.Context, O) (Undo, error)) (undo Undo, err error) {
	var undos []Undo
	undo = Undos(&undos)
	defer ExecUndo(undo, &err)
	for _, item := range items {
		u, e := itemFn(ctx, item)
		if err = e; e != nil {
			return
		}
		undos = append(undos, u)
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
		slog.LogAttrs(ctx, slog.LevelDebug, name+" done", attrs...)
	}
	return utils.Iif(optional, nil, err)
}

func ExecUndo(undo Undo, err *error) {
	if err == nil || *err != nil && undo != nil {
		undo()
	}
}

func Undos(undos *[]Undo) (undo Undo) {
	return func() {
		if undos == nil || len(*undos) == 0 {
			return
		}
		for _, undo := range slices.Backward(*undos) {
			if undo != nil {
				undo()
			}
		}
	}
}
