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
	"github.com/cnk3x/xunlei/pkg/utils"
)

type LinkOptions struct {
	Target   string
	Source   string
	DirMode  fs.FileMode
	Optional bool
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
	bq := utils.BackQueue(&undo, &err)
	defer bq.ErrDefer()

	var real string
	var act string
	err = func() (err error) {
		if real, err = filepath.EvalSymlinks(m.Source); err != nil {
			return
		}

		du, e := Mkdir(ctx, filepath.Dir(m.Target), m.DirMode)
		if err = e; err != nil {
			return
		}
		bq.Put(du)

		if err = os.Link(real, m.Target); !errors.Is(err, syscall.EXDEV) {
			act = "link"
			if err == nil {
				bq.Put(newRm(ctx, m.Target, "unlink"))
			}
			return
		}

		act = "copy"
		if err = fo.OpenRead(real, fo.ToFile(m.Target, fo.Perm(0777), fo.FlagExcl(false))); err == nil {
			bq.Put(newRm(ctx, m.Target, "uncopy"))
		}
		return
	}()

	attrs := []slog.Attr{
		slog.String("target", m.Target),
		slog.String("source", m.Source),
		slog.String("source_real", real),
		slog.Bool("optional", m.Optional),
	}

	if os.IsExist(err) {
		attrs = append(attrs, slog.String("err", os.ErrExist.Error()))
		slog.LogAttrs(ctx, slog.LevelDebug, act+" skip", attrs...)
		err = nil
	} else if os.IsNotExist(err) {
		attrs = append(attrs, slog.String("err", os.ErrNotExist.Error()))
		slog.LogAttrs(ctx, slog.LevelDebug, act+" skip", attrs...)
		err = nil
	} else {
		if err != nil {
			attrs = append(attrs, slog.String("err", err.Error()))
		}
		logIt(ctx, err, m.Optional, act, attrs...)
	}

	return
}

func Symlinks(ctx context.Context, items []LinkOptions) (undo Undo, err error) {
	return doMulti(ctx, items, Symlink)
}

func Symlink(ctx context.Context, m LinkOptions) (undo Undo, err error) {
	bq := utils.BackQueue(&undo, &err)
	defer bq.ErrDefer()

	err = func() (err error) {
		var dirUndo Undo
		if dirUndo, err = Mkdir(ctx, filepath.Dir(m.Target), m.DirMode); err != nil {
			return
		}
		bq.Put(dirUndo)

		if err = os.Symlink(m.Source, m.Target); err != nil {
			return
		}
		bq.Put(newRm(ctx, m.Target, "unlink"))

		return
	}()

	attrs := []slog.Attr{
		slog.String("target", m.Target),
		slog.String("source", m.Source),
	}

	act := "symlink"

	if os.IsExist(err) {
		attrs = append(attrs, slog.String("err", os.ErrExist.Error()))
		slog.LogAttrs(ctx, slog.LevelDebug, act+" skip", attrs...)
		err = nil
	} else if os.IsNotExist(err) {
		attrs = append(attrs, slog.String("err", os.ErrNotExist.Error()))
		slog.LogAttrs(ctx, slog.LevelDebug, act+" skip", attrs...)
		err = nil
	} else {
		if err != nil {
			attrs = append(attrs, slog.String("err", err.Error()))
		}
		logIt(ctx, err, m.Optional, act, attrs...)
	}

	return
}
