package sys

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/fo"
)

type LinkOptions struct {
	Target   string
	Source   string
	DirMode  fs.FileMode
	Optional bool
	Copy     bool
}

func Links(ctx context.Context, links []LinkOptions) (undo Undo, err error) {
	var undos []Undo
	undo = Undos(&undos)
	defer ExecUndo(undo, &err)
	for _, m := range links {
		u, e := Link(ctx, m)
		if err = e; e != nil {
			return
		}
		undos = append(undos, u)
	}
	return
}

// link hard link, only for file
func Link(ctx context.Context, m LinkOptions) (undo Undo, err error) {
	var undos []Undo
	undo = Undos(&undos)
	defer ExecUndo(undo, &err)

	real, e := filepath.EvalSymlinks(m.Source)
	if err = e; err != nil {
		return
	}

	dirUndo, e := Mkdir(ctx, filepath.Dir(m.Target), m.DirMode)
	if err = e; err != nil {
		return
	}
	undos = append(undos, dirUndo)

	err = os.Link(real, m.Target)
	if errors.Is(err, syscall.EXDEV) {
		err = fo.OpenRead(real, func(src *os.File) (err error) {
			return fo.OpenWrite(m.Target, fo.From(src), fo.PermFrom(src), fo.FlagExcl)
		})
	}

	attrs := []slog.Attr{
		slog.String("target", m.Target),
		slog.String("source", m.Source),
		slog.String("source_real", real),
		slog.Bool("optional", m.Optional),
	}

	if err != nil {
		attrs = append(attrs, slog.String("err", err.Error()))
	}

	switch {
	case os.IsExist(err):
		slog.LogAttrs(ctx, slog.LevelDebug, "link skip", attrs...)
	case err != nil && m.Optional:
		slog.LogAttrs(ctx, slog.LevelDebug, "link skip", attrs...)
	case err != nil && !m.Optional:
		slog.LogAttrs(ctx, slog.LevelWarn, "link fail", attrs...)
	default:
		slog.LogAttrs(ctx, slog.LevelDebug, "link done", attrs...)
		undos = append(undos, newRm(ctx, m.Target, "unlink"))
	}

	if os.IsExist(err) {
		err = nil
	}
	return
}
