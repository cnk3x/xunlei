package xlp

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Rotate 可依据文件大小进行分割，可指定是否压缩(gzip)
func Rotate(path string, maxsize string, compress bool) io.WriteCloser {
	f := &rotateFile{path: path, compress: compress}
	f.setMaxsize(maxsize)

	if err := f.open(); err != nil {
		fmt.Printf("[WARN] create rotate file: %v", err)
		return nopcw{io.Discard}
	}
	return f
}

type rotateFile struct {
	path     string // file path
	maxsize  uint64 // rotate file when gatter than this value
	compress bool   // compress backup file

	size uint64
	file *os.File

	mu sync.Mutex
}

func (f *rotateFile) open() (err error) {
	if err = os.MkdirAll(filepath.Dir(f.path), os.ModePerm); err != nil {
		return
	}

	var (
		file *os.File
		info os.FileInfo
	)

	if file, err = os.OpenFile(f.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		return
	}

	if info, err = file.Stat(); err != nil {
		return
	}

	f.file = file
	f.size = uint64(info.Size())
	return
}

func (f *rotateFile) rotate() (err error) {
	if err = f.close(); err != nil {
		return
	}

	if err = f.backup(f.path); err != nil {
		return
	}

	return f.open()
}

func (f *rotateFile) close() (err error) {
	if f.file != nil {
		if err = f.file.Sync(); err != nil {
			return
		}
		if err = f.file.Close(); err != nil {
			return
		}
		f.file = nil
	}
	return
}

func (f *rotateFile) Write(p []byte) (n int, err error) {
	if f.maxsize > 0 {
		f.mu.Lock()
		defer f.mu.Unlock()

		writeSize := uint64(len(p))
		if f.size+writeSize > f.maxsize {
			_ = f.rotate()
		}
	}

	n, err = f.file.Write(p)
	f.size += uint64(n)

	return
}

func (f *rotateFile) Close() (err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	err = f.close()
	return
}

func (f *rotateFile) backup(srcPath string) (err error) {
	ext := filepath.Ext(srcPath)
	dstPath := strings.TrimSuffix(srcPath, ext) + "-" + time.Now().Format("20060102-150405") + ext

	if f.compress {
		var src, dst *os.File
		if src, err = os.Open(srcPath); err != nil {
			return
		}

		if dst, err = os.Create(dstPath + ".gz"); err != nil {
			src.Close()
			return
		}

		z := gzip.NewWriter(dst)
		_, err = io.Copy(z, src)

		err = nSelect(err, z.Close(), dst.Close(), src.Close())

		if err == nil {
			os.RemoveAll(srcPath)
		} else {
			os.Rename(srcPath, dstPath)
		}
	} else {
		err = os.Rename(dstPath, dstPath)
	}

	return
}

func (f *rotateFile) setMaxsize(s string) {
	if s = strings.TrimSpace(s); s != "" {
		const units = "KMGT"
		var u float64
		if i := strings.IndexByte(units, s[len(s)-1]); i != -1 {
			s = s[:len(s)-1]
			u = float64(uint64(1) << ((i + 1) * 10))
		}

		var err error
		var n float64
		if n, err = strconv.ParseFloat(s, 64); err != nil {
			fmt.Printf("[WARN] %q is not a valid filesize", s)
			return
		}

		if u > 0 {
			n *= u
		}
		f.maxsize = uint64(n)

		// var formatSize = func(size uint64) string {
		// 	const units = "KMGT"
		// 	for i := len(units) - 1; i >= 0; i-- {
		// 		if u := uint64(1 << ((i + 1) * 10)); size >= u {
		// 			return fmt.Sprintf("%.3f%c", float64(size)/float64(u), units[i])
		// 		}
		// 	}
		// 	return fmt.Sprintf("%d", size)
		// }

		// fmt.Printf("logfile %s maxsize: %s -- %f -- %d -- %s", f.path, s, u, f.maxsize, formatSize(f.maxsize))
	}
}

type nopcw struct{ io.Writer }

func (nopcw) Close() (err error) { return }

func nSelect[T comparable](v T, f ...T) (out T) {
	if v == out {
		for _, f := range f {
			if f != out {
				return f
			}
		}
	}
	return v
}
