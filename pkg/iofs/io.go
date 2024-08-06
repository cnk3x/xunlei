package iofs

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

var buffer32KPool = &sync.Pool{New: func() interface{} { s := make([]byte, 32*1024); return &s }}

func WriteFileContext(ctx context.Context, src io.Reader, dstPath string, perm fs.FileMode) (err error) {
	if err = os.MkdirAll(filepath.Dir(dstPath), 0o777); err != nil {
		return
	}

	var dst *os.File
	if dst, err = os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm); err != nil {
		return
	}

	defer func() {
		if e := dst.Close(); e != nil && err == nil {
			err = e
		}
	}()

	if err = CopyContext(ctx, dst, src); err != nil {
		return
	}

	if perm != 0666 {
		err = dst.Chmod(perm)
	}

	return
}

func CopyContext(ctx context.Context, w io.Writer, r io.Reader) (err error) {
	buf := buffer32KPool.Get().(*[]byte)
	_, err = io.CopyBuffer(WriterContextWrap(ctx, w), ReaderContextWrap(ctx, r), *buf)
	buffer32KPool.Put(buf)
	return
}

func IORE(r io.Reader, err error) io.ReadCloser {
	if err == nil {
		if rc, ok := r.(io.ReadCloser); ok {
			return rc
		}
		return io.NopCloser(r)
	}
	return io.NopCloser(ioRw(func(p []byte) (int, error) { return 0, err }))
}

func WriterContextWrap(ctx context.Context, w io.Writer) io.Writer {
	return ioRw(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return w.Write(p)
		}
	})
}

func ReaderContextWrap(ctx context.Context, r io.Reader) io.Reader {
	return ioRw(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return r.Read(p)
		}
	})
}

func ErrRw(err error) io.ReadWriteCloser {
	return ioRw(func(p []byte) (int, error) { return 0, err })
}

type ioRw func(p []byte) (int, error)

func (fn ioRw) Read(p []byte) (int, error)  { return fn(p) }
func (fn ioRw) Write(p []byte) (int, error) { return fn(p) }
func (fn ioRw) Close() error                { return nil }

func NocWriter(src io.Writer) io.WriteCloser {
	return ioRw(func(p []byte) (int, error) { return src.Write(p) })
}
