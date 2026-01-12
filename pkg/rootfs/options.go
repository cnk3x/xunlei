package rootfs

import (
	"cmp"
	"context"
	"io/fs"
	"path/filepath"
)

const Separator = string(filepath.Separator)

// func Force(force bool) Option         { return func(ro *RunOptions) { ro.force = force } }
func After(after func() error) Option { return func(ro *RunOptions) { ro.after = after } }
func Before(before func(ctx context.Context) error) Option {
	return func(ro *RunOptions) { ro.before = before }
}

func Link(source, target string, dirMode fs.FileMode, optional ...bool) Option {
	return func(ro *RunOptions) {
		ro.links = append(ro.links, LinkOptions{
			Target:   target,
			Source:   []string{source},
			DirMode:  dirMode,
			Optional: cmp.Or(optional...),
		})
	}
}

func LinkRoot(source, newRoot string, dirMode fs.FileMode, optional ...bool) Option {
	return Link(source, filepath.Join(newRoot, source), dirMode, optional...)
}

func MountRoot(source, newRoot string, options ...MountOption) Option {
	return func(ro *RunOptions) {
		ro.mounts = append(ro.mounts, Mount(filepath.Join(newRoot, source), options...))
	}
}

func MountBindRoot(source, newRoot string, options ...MountOption) Option {
	return func(ro *RunOptions) {
		ro.mounts = append(ro.mounts,
			Mount(
				filepath.Join(newRoot, source),
				append(options, MountSource(source), MountBind())...,
			),
		)
	}
}

func Basic(ro *RunOptions) {
	optional := MountOptional(true)

	ro.mounts = append(ro.mounts,
		Mount(filepath.Join(ro.root, "proc"), MountType("proc")),                                                                      // 挂载proc文件系统（必须）
		Mount(filepath.Join(ro.root, "dev"), MountType("devtmpfs", MountDataMode("0755")), MountType("tmpfs", MountDataMode("0755"))), // 挂载devtmpfs到/dev（提供基础设备节点）
		Mount(filepath.Join(ro.root, "sys"), MountType("sysfs"), optional),                                                            // 挂载sysfs文件系统（可选，但建议挂载）
		Mount(filepath.Join(ro.root, "tmp"), MountType("tmpfs", MountDataSize("100m"), MountDataMode("0777")), optional),              // 挂载tmpfs到/tmp（临时目录，可选）
	)
}
