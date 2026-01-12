package fo

import (
	"cmp"
	"io"
	"os"
	"path/filepath"

	"github.com/cnk3x/xunlei/pkg/utils"
)

type Process func(*os.File) (err error)

func openFile(name string, process Process, defaultFlag int, options ...Option) (err error) {
	opts := Options{flag: defaultFlag, perm: 0666}

	for _, option := range options {
		option(&opts)
	}

	if opts.flag&os.O_CREATE != 0 {
		if err = os.MkdirAll(filepath.Dir(name), cmp.Or(opts.dirPerm, 0777)); err != nil {
			return
		}
	}

	f, e := os.OpenFile(name, opts.flag, cmp.Or(opts.perm, 0666))
	if err = e; err != nil {
		return
	}

	err = process(f)

	if e := f.Close(); e != nil && err == nil {
		err = e
	}
	return
}

func OpenWrite(name string, process Process, options ...Option) (err error) {
	return openFile(name, process, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, options...)
}

func OpenRead(name string, process Process, options ...Option) (err error) {
	return openFile(name, process, os.O_RDONLY, options...)
}

func To(w io.Writer) Process   { return func(r *os.File) error { return utils.Eol(io.Copy(w, r)) } }
func From(r io.Reader) Process { return func(w *os.File) error { return utils.Eol(io.Copy(w, r)) } }

func ToFile(to string, options ...Option) func(src *os.File) error {
	return func(src *os.File) error { return OpenWrite(to, From(src), options...) }
}

func FromFile(from string, options ...Option) func(w *os.File) error {
	return func(w *os.File) error { return OpenRead(from, To(w), options...) }
}

func Content[T ~string | ~[]byte](content T) func(w *os.File) error {
	return func(w *os.File) (err error) { return utils.Eol(w.Write([]byte(content))) }
}

func Lines[T ~string | ~[]byte](lines ...T) func(w *os.File) error {
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

func Nop(w *os.File) error { return nil }

// func Eol[T any](_ T, err error) error { return err }
// func Eon[T any, E any](v T, _ E) T    { return v }
