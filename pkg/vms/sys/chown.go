package sys

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

func Chown(ctx context.Context, paths []string, uid, gid uint32, recursive bool) {
	_chown := func(path string, uid, gid uint32) {
		if err := os.Chown(path, int(uid), int(gid)); err != nil {
			slog.DebugContext(ctx, "chown", "uid", uid, "gid", gid, "path", path)
		}
	}

	chown := func(path string, uid, gid uint32, recursive bool) {
		if recursive {
			filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				_chown(path, uid, gid)
				return nil
			})
		} else {
			_chown(path, uid, gid)
		}
	}

	for _, path := range paths {
		chown(path, uid, gid, recursive)
	}
}
