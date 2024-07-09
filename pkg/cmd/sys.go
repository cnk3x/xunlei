package cmd

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cnk3x/xunlei/pkg/lod"
)

func WalkDir(ctx context.Context, root string, fn func(path string, d fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return fn(path, d)
		}
	})
}

func Chown(ctx context.Context, path string, uid, gid uint32, recursion ...bool) (err error) {
	chown := func(p string, uid, gid uint32) (err error) {
		err = os.Chown(p, int(uid), int(gid))
		slog.Log(ctx, lod.ErrDebug(err), "chown", "uid", uid, "gid", gid, "path", p, "err", err)
		return
	}

	if lod.First(recursion) {
		return WalkDir(ctx, path, func(p string, _ fs.DirEntry) error { return chown(p, uid, gid) })
	}

	return chown(path, uid, gid)
}

func Chmod(ctx context.Context, path string, mode fs.FileMode, recursion ...bool) (err error) {
	chmod := func(path string, mode fs.FileMode) (err error) {
		err = os.Chmod(path, mode)
		slog.Log(ctx, lod.ErrDebug(err), "chmod", "mode", mode, "path", path, "err", err)
		return
	}

	if lod.First(recursion) {
		return WalkDir(ctx, path, func(path string, _ fs.DirEntry) error { return chmod(path, mode) })
	}

	return chmod(path, mode)
}
