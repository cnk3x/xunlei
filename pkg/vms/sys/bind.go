package sys

import (
	"context"
	"log/slog"
	"path/filepath"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/utils"
)

type BindOptions struct {
	Source   string
	Target   string
	Optional bool
}

func Binds(ctx context.Context, items []BindOptions) (undo Undo, err error) {
	return doMulti(ctx, items, Bind)
}

func BindsTask(ctx context.Context, items ...BindOptions) func() (undo Undo, err error) {
	return func() (undo Undo, err error) { return Binds(ctx, items) }
}

func BindRs(ctx context.Context, root string, items []BindOptions) (undo Undo, err error) {
	for i := range items {
		if items[i].Target == "" {
			items[i].Target = filepath.Join(root, items[i].Source)
		}
	}
	return doMulti(ctx, items, Bind)
}

func BindRsTask(ctx context.Context, root string, items ...BindOptions) func() (undo Undo, err error) {
	return func() (undo Undo, err error) { return BindRs(ctx, root, items) }
}

// 绑定文件夹
func Bind(ctx context.Context, m BindOptions) (undo Undo, err error) {
	bq := utils.BackQueue(&undo, &err)
	defer bq.ErrDefer()

	real, e := filepath.EvalSymlinks(m.Source)
	err = func() (err error) {
		if err = e; err != nil {
			return
		}

		u, e := Mkdir(ctx, m.Target, 0o777, true)
		if err = e; err != nil {
			return
		}
		bq.Put(u)

		if err = syscall.Mount(real, m.Target, "", syscall.MS_BIND, ""); err != nil {
			return
		}
		bq.Put(mkUnmount(ctx, m.Target, "unbind"))
		return
	}()

	err = checkErr(ctx, err, m.Optional, "bind",
		slog.String("target", m.Target),
		slog.String("source", m.Source),
		slog.String("source_real", real),
	)
	return
}
