package spk

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/cnk3x/xunlei/pkg/iofs"
	"github.com/cnk3x/xunlei/pkg/lod"
	"github.com/ulikunitz/xz"
)

// ExtractEmbedSpk 从迅雷SPK中提取需要的文件
func ExtractEmbedSpk(ctx context.Context, dstDir string) error {
	return ExtractSpk(ctx, bytes.NewReader(Bytes), dstDir)
}

// ExtractSpkFile 从迅雷SPK中提取需要的文件
func ExtractSpkFile(ctx context.Context, srcPath string, dstDir string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()
	return ExtractSpk(ctx, src, dstDir)
}

// ExtractSpk 从迅雷SPK中提取需要的文件
func ExtractSpk(ctx context.Context, src io.Reader, dstDir string) (err error) {
	return TarExtract(ctx, src, func(ctx context.Context, tr io.Reader, h *tar.Header) (err error) {
		if h.Name == "package.tgz" {
			if err = ExtractSpkPackage(ctx, tr, dstDir); err != nil {
				return
			}
			err = fs.SkipAll
		}
		return
	})
}

// ExtractSpkPackage 从迅雷SPK文件内的package.tgz中提取需要的文件
func ExtractSpkPackage(ctx context.Context, src io.Reader, dstDir string) error {
	return TarExtract(ctx, src, func(ctx context.Context, tr io.Reader, h *tar.Header) (err error) {
		var perm fs.FileMode

		switch {
		case strings.HasPrefix(h.Name, "bin/bin/version"):
			perm = 0666
		case strings.HasPrefix(h.Name, "bin/bin/xunlei-pan-cli"):
			perm = 0777
		case h.Name == "ui/index.cgi":
			perm = 0777
		default:
			return
		}

		err = iofs.WriteFileContext(ctx, tr, filepath.Join(dstDir, h.Name), perm)
		slog.Log(ctx, lod.ErrDebug(err), "extract package", "perm", perm, "target_dir", dstDir, "name", h.Name, "err", err)
		return
	}, Xz)
}

func Xz(src io.Reader) (io.ReadCloser, error) {
	xzr, err := xz.NewReader(src)
	return io.NopCloser(xzr), err
}

/* tar decode functions */

type TarDecoder func(io.Reader) (io.ReadCloser, error)

func TarExtract(ctx context.Context, src io.Reader, walk TarWalkFunc, decoder ...TarDecoder) (err error) {
	dr := lod.First(decoder).Decode(src)
	defer dr.Close()

	var hdr *tar.Header
	tr := tar.NewReader(dr)
	for hdr, err = tr.Next(); err != io.EOF; hdr, err = tr.Next() {
		if err != nil {
			return
		}
		err = walk.Read(ctx, tr, hdr)
	}

	if errors.Is(err, fs.SkipAll) || err == io.EOF {
		err = nil
	}
	return
}

func (d TarDecoder) Decode(r io.Reader) io.ReadCloser {
	if d != nil {
		if rc, err := d(r); err == nil {
			return rc
		} else {
			return iofs.ErrRw(err)
		}
	}
	return io.NopCloser(r)
}

type TarWalkFunc func(ctx context.Context, r io.Reader, h *tar.Header) (err error)

func (w TarWalkFunc) Read(ctx context.Context, r io.Reader, h *tar.Header) (err error) {
	defer io.Copy(io.Discard, r)
	if w != nil {
		return w(ctx, r, h)
	}
	return nil
}
