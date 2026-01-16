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

	var undos []Undo
	undo = Undos(&undos)
	defer ExecUndo(undo, &err)

	if src, err = filepath.EvalSymlinks(src); err == nil {
		var dirUndo Undo
		if dirUndo, err = Mkdir(ctx, m.Target, 0777); err == nil {
			undos = append(undos, dirUndo)
			if err = syscall.Mount(src, m.Target, "", syscall.MS_BIND, ""); err == nil {
				undos = append(undos, mkUnmount(ctx, m.Target, "unbind"))
			}
		}
	}

	attrs := []slog.Attr{
		slog.String("target", m.Target),
		slog.String("source", m.Source),
		slog.String("source_real", src),
	}

	err = logIt(ctx, err, m.Optional, "bind", attrs...)
	return
}
