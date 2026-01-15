package rootfs

import (
	"cmp"
	"context"
	"io/fs"
	"path/filepath"

	"github.com/cnk3x/xunlei/pkg/rootfs/sys"
)

func After(after func() error) Option { return func(ro *RunOptions) { ro.after = after } }

func Before(before func(ctx context.Context) error) Option {
	return func(ro *RunOptions) { ro.before = before }
}

func Link(source, target string, dirMode fs.FileMode, optional ...bool) Option {
	return func(ro *RunOptions) {
		ro.links = append(ro.links, sys.LinkOptions{
			Target:   target,
			Source:   source,
			DirMode:  dirMode,
			Optional: cmp.Or(optional...),
		})
	}
}

func LinkRoot(source, newRoot string, dirMode fs.FileMode, optional ...bool) Option {
	return Link(source, filepath.Join(newRoot, source), dirMode, optional...)
}

func Basic(ro *RunOptions) {
	ro.mounts = append(ro.mounts,
		// 挂载proc文件系统（必须）
		sys.MountOptions{Target: filepath.Join(ro.root, "proc"), Source: "proc", Fstype: "proc"},
		// 挂载devtmpfs到/dev（提供基础设备节点）
		sys.MountOptions{Target: filepath.Join(ro.root, "dev"), Source: "devtmpfs", Fstype: "devtmpfs", Data: "mode=0755"}, //tmpfs
		// 挂载sysfs文件系统（可选，但建议挂载）
		sys.MountOptions{Target: filepath.Join(ro.root, "sys"), Source: "sysfs", Fstype: "sysfs", Optional: true},
		// 挂载tmpfs到/tmp（临时目录，可选）
		sys.MountOptions{Target: filepath.Join(ro.root, "tmp"), Source: "tmpfs", Fstype: "tmpfs", Data: "mode=0777,size=100m", Optional: true},
	)
}

func Links(root string, files ...string) Option {
	return func(ro *RunOptions) {
		for _, file := range files {
			mOpts := sys.LinkOptions{
				Target:   filepath.Join(root, file),
				Source:   file,
				Optional: true,
				DirMode:  0777,
			}
			ro.links = append(ro.links, mOpts)
		}
	}
}

func Binds(root string, dirs ...string) func(ro *RunOptions) {
	return func(ro *RunOptions) {
		for _, dir := range dirs {
			mOpts := sys.BindOptions{
				Target:   filepath.Join(root, dir),
				Source:   dir,
				Optional: true,
			}
			ro.binds = append(ro.binds, mOpts)
		}
	}
}
