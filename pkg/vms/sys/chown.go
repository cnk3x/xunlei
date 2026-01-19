package sys

import (
	"cmp"
	"context"
	"io/fs"
	"log/slog"
	"os"

	"github.com/cnk3x/xunlei/pkg/fo"
)

// Chown 更改指定文件的数字用户（uid）和组（gid）
//   - 与标准库 os.Chown 行为不同，如果uid或者gid为0(即 root)，不会做改变，也就是说，该方法无法将用户和组改成root
func Chown(ctx context.Context, paths []string, uid, gid int, recursive bool) {
	uid, gid = cmp.Or(uid, -1), cmp.Or(uid, -1)

	if uid == -1 && gid == -1 {
		return
	}

	chown := func(path string, _ fs.DirEntry) error {
		if err := os.Chown(path, uid, gid); err != nil {
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
