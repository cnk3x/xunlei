package vms

import (
	"context"
	"path/filepath"

	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/cnk3x/xunlei/pkg/vms/sys"
)

func Root(root string) Option { return func(ro *options) { ro.root = root } }

func Run(run func(ctx context.Context) error) Option { return func(ro *options) { ro.run = run } }

func User[U, G utils.IntT | utils.UintT](uid U, gid G) Option {
	return func(ro *options) { ro.uid, ro.gid = int(uid), int(gid) }
}

func Debug(debug ...bool) Option {
	return func(ro *options) { ro.debug = len(debug) == 0 || debug[0] }
}

func After(after func(ctx context.Context, runErr error) (err error)) Option {
	return func(ro *options) { ro.after = after }
}

func Before(before func(ctx context.Context) (undo func(), err error)) Option {
	return func(ro *options) { ro.before = before }
}

func Basic(ro *options) {
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

func Links(files ...string) Option {
	return func(ro *options) {
		for _, file := range files {
			mOpts := sys.LinkOptions{
				Source:   file,
				Optional: true,
				DirMode:  0777,
			}
			ro.links = append(ro.links, mOpts)
		}
	}
}

func Binds(dirs ...string) func(ro *options) {
	return func(ro *options) {
		for _, dir := range dirs {
			mOpts := sys.BindOptions{
				Source:   dir,
				Optional: true,
			}
			ro.binds = append(ro.binds, mOpts)
		}
	}
}
