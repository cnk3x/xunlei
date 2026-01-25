package fo

import (
	"io/fs"
	"os"
)

// options 创建文件时候的选项
type options struct {
	flag    int
	perm    fs.FileMode
	dirPerm fs.FileMode
	existOk bool
}

// Option 创建文件时候的选项
type Option func(*options)

// Perm 创建文件时候的权限
func Perm(perm fs.FileMode) Option { return func(o *options) { o.perm = perm } }

// DirPerm 创建文件时，如果上级目录不存在，会自动创建，该值指向创建这些目录时所设置的权限
func DirPerm(perm fs.FileMode) Option { return func(o *options) { o.dirPerm = perm } }

// Flag 创建文件时候的 flag
func Flag(flag int) Option { return func(o *options) { o.flag = flag } }

// FlagExcl 创建文件时候如果文件已存在则返回错误，如果同时 existOK 为 true 则忽略该错误，且什么都不会做。
func FlagExcl(existOk bool) Option {
	return func(o *options) {
		o.flag, o.existOk = o.flag&^fwo|os.O_EXCL, existOk
	}
}

// 创建文件时候的 flag, 追加模式
func FlagAppend(o *options) { o.flag = o.flag&^fwo | os.O_APPEND }

// 创建文件时候的 flag, 覆盖模式
func FlagTrunc(o *options) { o.flag = o.flag&^fwo | os.O_TRUNC }

// 创建文件时候的 flag, 只写模式(write only)
func FlagWo(o *options) { o.flag = o.flag&^frw | os.O_WRONLY }

// 创建文件时候的 flag, 读写模式 (read and write)
func FlagRw(o *options) { o.flag = o.flag&^frw | os.O_RDWR }

// 从文件中获取权限信息，设置到新文件中，只在新创建的文件中生效
func PermFrom(r *os.File) Option {
	return func(o *options) {
		if stat, err := r.Stat(); err == nil {
			o.perm = stat.Mode().Perm()
		}
	}
}

const fwo = os.O_APPEND | os.O_TRUNC | os.O_EXCL
const frw = os.O_RDWR | os.O_WRONLY

// FlagRepl 从 src 中清理 mask 中指定的 flag, 替换成 new (src&^mask | new)
func FlagRepl(src, mask, new int) int { return src&^mask | new }
