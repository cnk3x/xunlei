package sys

import (
	"context"
	"log/slog"
	"path/filepath"
	"syscall"
)

type BindOptions struct {
	Source   string
	Target   string
	Optional bool
}

func Binds(ctx context.Context, items []BindOptions) (undo Undo, err error) {
	return doMulti(ctx, items, Bind)
}

// 绑定文件夹
func Bind(ctx context.Context, m BindOptions) (undo Undo, err error) {
	src := m.Source
	if src, err = filepath.EvalSymlinks(src); err == nil {
		err = syscall.Mount(src, m.Target, "", syscall.MS_BIND, "")
	}

	if err == nil {
		undo = mkUnmount(ctx, m.Target, "unbind")
	}

	attrs := []slog.Attr{
		slog.String("target", m.Target),
		slog.String("source", m.Source),
		slog.String("source_real", src),
	}

	err = logIt(ctx, err, m.Optional, "bind", attrs...)
	return
}
