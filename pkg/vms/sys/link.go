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

	act := "link"
	real, e := filepath.EvalSymlinks(m.Source)
	err = func() (err error) {
		if err = e; err != nil {
			return
		}

		u, e := Mkdir(ctx, filepath.Dir(m.Target), m.DirMode)
		if err = e; err != nil {
			return
		}
		bq.Put(u)

		if err = os.Link(real, m.Target); !errors.Is(err, syscall.EXDEV) {
			if err == nil {
				bq.Put(newRm(ctx, m.Target, "unlink"))
			}
			return
		}

		act = "copy"
		err = fo.OpenRead(real, fo.ToFile(m.Target, fo.Perm(0o777), fo.FlagExcl(false)))
		if err != nil {
			return
		}
		bq.Put(newRm(ctx, m.Target, "uncopy"))
		return
	}()

	err = checkErr(ctx, err, m.Optional, act,
		slog.String("target", m.Target),
		slog.String("source", m.Source),
		slog.String("source_real", real),
		slog.Bool("optional", m.Optional),
	)
	return
}

func Symlinks(ctx context.Context, items []LinkOptions) (undo Undo, err error) {
	return doMulti(ctx, items, Symlink)
}

func Symlink(ctx context.Context, m LinkOptions) (undo Undo, err error) {
	bq := utils.BackQueue(&undo, &err)
	defer bq.ErrDefer()

	err = func() (err error) {
		u, e := Mkdir(ctx, filepath.Dir(m.Target), m.DirMode)
		if err = e; err != nil {
			return
		}
		bq.Put(u)

		if err = os.Symlink(m.Source, m.Target); err != nil {
			return
		}
		bq.Put(newRm(ctx, m.Target, "rmSymlink"))
		return
	}()

	err = checkErr(ctx, err, m.Optional, "symlink",
		slog.String("target", m.Target),
		slog.String("source", m.Source),
	)
	return
}
