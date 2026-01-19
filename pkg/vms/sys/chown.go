package sys

import (
	"context"
	"io/fs"
	"log/slog"
	"os"

	"github.com/cnk3x/xunlei/pkg/fo"
)

func Chown(ctx context.Context, paths []string, uid, gid uint32, recursive bool) {
	chown := func(path string, _ fs.DirEntry) error {
		if err := os.Chown(path, int(uid), int(gid)); err != nil {
			slog.DebugContext(ctx, "chown", "uid", uid, "gid", gid, "path", path, "err", err)
		} else {
			slog.DebugContext(ctx, "chown", "uid", uid, "gid", gid, "path", path)
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
