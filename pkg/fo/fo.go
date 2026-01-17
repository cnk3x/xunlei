package fo

import (
	"cmp"
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cnk3x/xunlei/pkg/utils"
)

type Process func(*os.File) (err error)

func openFile(name string, process Process, defaultFlag int, fOpts ...Option) (err error) {
	opts := options{flag: defaultFlag}
	for _, apply := range fOpts {
		apply(&opts)
	}

	if opts.flag&os.O_CREATE != 0 {
		if err = os.MkdirAll(filepath.Dir(name), cmp.Or(opts.dirPerm, 0777)); err != nil {
			return
		}
	}

	f, e := os.OpenFile(name, opts.flag, cmp.Or(opts.perm, 0666))
	if e != nil && !(opts.existOk && os.IsExist(e)) {
		err = e
		return
	}

	if err, e = process(f), f.Close(); err == nil && e != nil {
		err = e
	}
	return
}

// OpenWrite 写入
func OpenWrite(name string, process Process, options ...Option) (err error) {
	return openFile(name, process, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, options...)
}

// OpenRead 读取
func OpenRead(name string, process Process, options ...Option) (err error) {
	return openFile(name, process, os.O_RDONLY, options...)
}

// To 将读取的文件写入到w
func To(w io.Writer) Process { return func(r *os.File) error { return utils.Eol(io.Copy(w, r)) } }

// From 将源r写入到文件
func From(r io.Reader) Process { return func(w *os.File) error { return utils.Eol(io.Copy(w, r)) } }

// To File 将读取的文件写入到文件
func ToFile(to string, options ...Option) Process {
	return func(src *os.File) error { return OpenWrite(to, From(src), options...) }
}

// FromFile 从文件读取并写入到打开的文件
func FromFile(from string, options ...Option) Process {
	return func(w *os.File) error { return OpenRead(from, To(w), options...) }
}

// 将内容写入到打开的文件
func Content[T ~string | ~[]byte](content T) Process {
	return func(w *os.File) (err error) { return utils.Eol(w.Write([]byte(content))) }
}

// 将内容行写入到打开的文件
func Lines[T ~string | ~[]byte](lines ...T) Process {
	return func(w *os.File) (err error) {
		for _, line := range lines {
			if _, err = w.Write([]byte(line)); err != nil {
				return
			}
			if len(lines) > 1 {
				if _, err = w.Write([]byte("\n")); err != nil {
					return
				}
			}
		}
		return
	}
}

// 什么都不干
func Nop(w *os.File) error { return nil }

func WriteFile(ctx context.Context, fn string, process Process, fopts ...Option) (undo func(), err error) {
	var exists bool
	undo = func() {
		if exists {
			if e := os.Remove(fn); e != nil {
				slog.DebugContext(ctx, "remove syno info", "err", e)
			}
		}
	}

	err = OpenWrite(fn, process, append(fopts, FlagExcl(true))...)
	if exists = err != nil && os.IsExist(err); exists {
		err = nil
	}

	return
}
