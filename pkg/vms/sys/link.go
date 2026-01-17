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

func Links(ctx context.Context, items []LinkOptions) (undo Undo, err error) {
	return doMulti(ctx, items, Link)
}

func LinkRs(ctx context.Context, root string, items []LinkOptions) (undo Undo, err error) {
	for i := range items {
		if items[i].Target == "" {
			items[i].Target = filepath.Join(root, items[i].Source)
		}
	}
	return doMulti(ctx, items, Link)
}

// link hard link, only for file
func Link(ctx context.Context, m LinkOptions) (undo Undo, err error) {
	var undos []Undo
	undo = Undos(&undos)
	defer ExecUndo(undo, &err)

	var real string
	err = func() (err error) {
		if real, err = filepath.EvalSymlinks(m.Source); err != nil {
			return
		}

		du, e := Mkdir(ctx, filepath.Dir(m.Target), m.DirMode)
		if err = e; err != nil {
			return
		}
		undos = append(undos, du)

		if err = os.Link(real, m.Target); !errors.Is(err, syscall.EXDEV) {
			return
		}

		err = fo.OpenRead(real, fo.ToFile(m.Target, fo.Perm(0777), fo.FlagExcl(false)))
		return
	}()

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
		err = nil
		slog.LogAttrs(ctx, slog.LevelDebug, "link skip", attrs...)
	case err != nil && m.Optional:
		slog.LogAttrs(ctx, slog.LevelDebug, "link skip", attrs...)
		err = nil
	case err != nil && !m.Optional:
		slog.LogAttrs(ctx, slog.LevelWarn, "link fail", attrs...)
	default:
		slog.LogAttrs(ctx, slog.LevelDebug, "link done", attrs...)
		undos = append(undos, newRm(ctx, m.Target, "unlink"))
	}

	return
}
