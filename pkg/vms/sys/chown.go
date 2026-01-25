package sys

import (
	"cmp"
	"context"
	"io/fs"
	"log/slog"
	"os"

	"github.com/cnk3x/xunlei/pkg/fo"
	"github.com/cnk3x/xunlei/pkg/utils"
)

// Chown 更改指定文件的数字用户（uid）和组（gid）
//   - 与标准库 os.Chown 行为不同，如果uid或者gid为0(即 root)，不会做改变，也就是说，该方法无法将用户和组改成root
func Chown(ctx context.Context, paths []string, uid, gid int, recursive bool) {
	uid, gid = cmp.Or(uid, -1), cmp.Or(uid, -1)

	if uid == -1 && gid == -1 {
		return
	}

	chown := func(path string, _ fs.DirEntry) error {
		if utils.Eol(os.Stat(path)) == nil {
			if err := os.Chown(path, uid, gid); err != nil {
				slog.DebugContext(ctx, "chown", "uid", uid, "gid", gid, "path", path, "err", err)
			} else {
				slog.DebugContext(ctx, "chown", "uid", uid, "gid", gid, "path", path)
			}
		}
		return nil
	}

	processItem := func(path string, recursive bool) {
		if recursive {
			fo.WalkDir(path, chown)
		} else {
			chown(path, nil)
		}
	}

	for _, path := range paths {
		processItem(path, recursive)
	}
}

// Chown 更改指定文件的数字用户（uid）和组（gid）
//   - 与标准库 os.Chown 行为不同，如果uid或者gid为0(即 root)，不会做改变，也就是说，该方法无法将用户和组改成root
func Chmod(ctx context.Context, paths []string, perm fs.FileMode, recursive bool) {
	chown := func(path string, _ fs.DirEntry) error {
		if utils.Eol(os.Stat(path)) == nil {
			if err := os.Chmod(path, perm); err != nil {
				slog.DebugContext(ctx, "chmod", "perm", Perm2s(perm), "path", path, "err", Errno(err))
			} else {
				slog.DebugContext(ctx, "chmod", "perm", Perm2s(perm), "path", path)
			}
		}
		return nil
	}

	processItem := func(path string, recursive bool) {
		if recursive {
			fo.WalkDir(path, chown)
		} else {
			chown(path, nil)
		}
	}

	for _, path := range paths {
		processItem(path, recursive)
	}
}

func ChmodTask(ctx context.Context, paths []string, perm fs.FileMode, recursive bool) func() (Undo, error) {
	return func() (undo Undo, err error) {
		Chmod(ctx, paths, perm, recursive)
		return
	}
}

func ChownTask(ctx context.Context, paths []string, uid, gid int, recursive bool) func() (Undo, error) {
	return func() (undo Undo, err error) {
		Chown(ctx, paths, uid, gid, recursive)
		return
	}
}
