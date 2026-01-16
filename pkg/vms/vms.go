package vms

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"

	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/vms/sys"
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
	debug  bool
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

	undo, e := seqExecWitUndo(
		func() (sys.Undo, error) { return sys.Mounts(ctx, opts.mounts) }, //mounts
		func() (sys.Undo, error) { return sys.Binds(ctx, opts.binds) },   //binds
		func() (sys.Undo, error) { return sys.Links(ctx, opts.links) },   //links
	)
	if err = e; err != nil {
		return
	}
	defer undo()

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

	if opts.root == "" || opts.root == "/" {
		err = run(ctx)
	} else {
		//chroot & run
		err = sys.Chroot(log.Prefix(ctx, "prog"), opts.root, run, opts.debug)
	}
	return
}

func seqExecWitUndo(fns ...func() (undo sys.Undo, err error)) (undo sys.Undo, err error) {
	var undos []sys.Undo
	undo = sys.Undos(&undos)
	defer sys.ExecUndo(undo, &err)

	for _, fn := range fns {
		u, e := fn()
		if err = e; err != nil {
			return
		}
		undos = append(undos, u)
	}

	return
}
