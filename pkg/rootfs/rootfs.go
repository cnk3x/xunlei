package rootfs

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"

	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/rootfs/sys"
	"github.com/samber/lo"
)

// type Undo func()
type Option func(ro *RunOptions)

type Options []Option

func (o Options) Add(options ...Option) Options { return append(o, options...) }

type RunOptions struct {
	root   string
	mounts []sys.MountOptions
	binds  []sys.BindOptions
	links  []sys.LinkOptions
	before func(ctx context.Context) error
	after  func() error
}

func Run(ctx context.Context, newRoot string, run func(ctx context.Context) error, options ...Option) (err error) {
	slog.InfoContext(ctx, "start boot")

	defer func() {
		if errors.Is(err, context.Canceled) {
			slog.InfoContext(ctx, "rootfs canceled")
			err = nil
		}
	}()

	var opts RunOptions
	if opts.root, err = filepath.Abs(newRoot); err != nil {
		return
	}

	for _, option := range options {
		option(&opts)
	}

	if opts.root == "" || opts.root == "/" {
		err = run(ctx)
		return
	}

	// //兜底
	// defer func() {
	// 	if opts.force {
	// 		exec.Command("sh", "-c", `mount | grep `+newRoot+` | grep on | awk '{print $3}' | xargs -I {} umount {}`).Run()
	// 	}
	// }()

	//undos
	var undos []sys.Undo
	defer sys.ExecUndo(sys.Undos(&undos), nil)

	var undo sys.Undo

	//mkdirs
	dirs := lo.Map(opts.mounts, func(m sys.MountOptions, _ int) string { return m.Target })
	if undo, err = sys.Mkdirs(ctx, dirs, 0777); err != nil {
		return
	}
	undos = append(undos, undo)

	//mounts
	if undo, err = sys.Mounts(ctx, opts.mounts); err != nil {
		return
	}
	undos = append(undos, undo)

	//links
	if undo, err = sys.Links(ctx, opts.links); err != nil {
		return
	}
	undos = append(undos, undo)

	//after
	defer func() {
		if opts.after != nil {
			if err = opts.after(); err != nil {
				return
			}
		}
	}()

	//before
	if opts.before != nil {
		if err = opts.before(ctx); err != nil {
			return
		}
	}

	//chroot & run
	err = sys.Chroot(log.Prefix(ctx, "prog"), opts.root, run)
	return
}
