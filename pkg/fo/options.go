package fo

import (
	"io/fs"
	"os"
)

type Options struct {
	flag    int
	perm    fs.FileMode
	dirPerm fs.FileMode
}

type Option func(*Options)

func Perm(perm fs.FileMode) Option    { return func(o *Options) { o.perm = perm } }
func DirPerm(perm fs.FileMode) Option { return func(o *Options) { o.dirPerm = perm } }
func Flag(flag int) Option            { return func(o *Options) { o.flag = flag } }

func FlagExcl(o *Options)   { o.flag = fr0(o.flag, os.O_EXCL) }
func FlagAppend(o *Options) { o.flag = fr0(o.flag, os.O_APPEND) }
func FlagTrunc(o *Options)  { o.flag = fr0(o.flag, os.O_TRUNC) }
func FlagRo(o *Options)     { o.flag = fr1(o.flag, os.O_WRONLY) }
func FlagRw(o *Options)     { o.flag = fr1(o.flag, os.O_RDWR) }

func PermFrom(r *os.File) Option {
	return func(o *Options) {
		if stat, err := r.Stat(); err == nil {
			o.perm = stat.Mode().Perm()
		}
	}
}

func fr0(src int, newFlag int) int { return src&^(os.O_APPEND|os.O_TRUNC|os.O_EXCL) | newFlag }
func fr1(src int, newFlag int) int { return src&^(os.O_RDWR|os.O_WRONLY) | newFlag }
