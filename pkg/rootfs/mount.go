package rootfs

import (
	"cmp"
	"fmt"
)

type MountOption func(*MountOptions)

func Mount(target string, options ...MountOption) MountOptions {
	mount := MountOptions{Target: target}
	for _, option := range options {
		option(&mount)
	}
	return mount
}

func MountBind(v ...bool) MountOption {
	return func(mpo *MountOptions) { mpo.Bind = len(v) == 0 || cmp.Or(v...) }
}

func MountOptional(v ...bool) MountOption {
	return func(mpo *MountOptions) { mpo.Optional = len(v) == 0 || cmp.Or(v...) }
}

func Optional(v ...bool) MountOption {
	return func(mpo *MountOptions) { mpo.Optional = len(v) == 0 || cmp.Or(v...) }
}

func Target(v string) MountOption { return func(mo *MountOptions) { mo.Target = v } }

type MountSourceOption func(*MountSourceOptions)

func MountSource(source string, options ...MountSourceOption) MountOption {
	s := MountSourceOptions{Source: source}
	for _, option := range options {
		option(&s)
	}
	return func(mpo *MountOptions) { mpo.Source = append(mpo.Source, s) }
}

func MountType(source string, options ...MountSourceOption) MountOption {
	return MountSource(source, append(options, MountFstype(source))...)
}

func MountFstype(v string) MountSourceOption { return func(ms *MountSourceOptions) { ms.Fstype = v } }
func MountFlags(v uintptr) MountSourceOption { return func(ms *MountSourceOptions) { ms.Flags = v } }
func MountData(v string) MountSourceOption {
	return func(mso *MountSourceOptions) {
		mso.Data += iif(mso.Data == "", "", ",") + v
	}
}

func MountDataMode(perm string) MountSourceOption { return MountData(fmt.Sprintf("mode=%s", perm)) }
func MountDataSize(size string) MountSourceOption { return MountData(fmt.Sprintf("size=%s", size)) }
