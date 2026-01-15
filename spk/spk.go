package spk

import (
	"archive/tar"
	"cmp"
	"context"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/ulikunitz/xz"
)

// Extract 从迅雷SPK中提取需要的文件(存在则跳过)
func Extract(ctx context.Context, src io.Reader, dstDir string) (err error) {
	return Walk(ctx, src, func(tr io.Reader, h *tar.Header) (err error) {
		if h.Name == "package.tgz" {
			err = cmp.Or(Walk(ctx, tr, func(tr io.Reader, h *tar.Header) (err error) {
				var perm fs.FileMode
				switch {
				case strings.HasPrefix(h.Name, "bin/bin/version"):
					perm = 0o666
				case strings.HasPrefix(h.Name, "bin/bin/xunlei-pan-cli"):
					perm = 0o777
				case h.Name == "ui/index.cgi":
					perm = 0o777
				default:
					return
				}

				err = func() (err error) {
					target := filepath.Join(dstDir, h.Name)
					if err = os.MkdirAll(filepath.Dir(target), 0o777); err != nil {
						return
					}

					f, e := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
					if e != nil {
						if !os.IsExist(e) {
							err = e
						}
						return
					}
					_, err = io.Copy(f, tr)
					if ce := f.Close(); err == nil && ce != nil {
						err = ce
					}
					return
				}()

				slog.Log(ctx, log.ErrDebug(err), "extract package", "perm", perm, "target_dir", dstDir, "name", h.Name, "err", err)
				return
			}, Xz), io.EOF)
		}
		return
	})
}

/* tar decode functions */

func Walk(ctx context.Context, src io.Reader, walk WalkFunc, decoder ...Decoder) (err error) {
	dr := io.NopCloser(src)

	for _, d := range decoder {
		if d != nil {
			if dr, err = d(dr); err != nil {
				return
			}
			defer dr.Close()
		}
	}

	for tr := tar.NewReader(dr); ; {
		hdr, e := tr.Next()
		if err = e; err != nil {
			break
		}
		if err = walk(tr, hdr); err != nil {
			break
		}
	}

	if err == io.EOF {
		err = nil
	}
	return
}

func Xz(src io.Reader) (io.ReadCloser, error) {
	xzr, err := xz.NewReader(src)
	return io.NopCloser(xzr), err
}

type Decoder func(io.Reader) (io.ReadCloser, error)

type WalkFunc func(r io.Reader, h *tar.Header) (err error)
