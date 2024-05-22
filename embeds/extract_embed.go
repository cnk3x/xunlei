//go:build embed

package embeds

import (
	"archive/tar"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

//go:embed nasxunlei.rpk
var rpkBytes []byte

func Extract(target string) (err error) {
	var d *zstd.Decoder
	if d, err = zstd.NewReader(bytes.NewReader(rpkBytes)); err != nil {
		return
	}

	extract := func(tr *tar.Reader, h *tar.Header) (err error) {
		info := h.FileInfo()
		dstPath := filepath.Join(target, h.Name)

		if err = os.MkdirAll(filepath.Dir(dstPath), 0777); err != nil {
			return
		}

		var fw *os.File
		if fw, err = os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666); err != nil {
			if os.IsExist(err) {
				slog.Warn(fmt.Sprintf("  - %s is exist. ignore %s", dstPath, h.Name))
				err = nil
			}
			return
		}
		defer fw.Close()

		// 将 tr 写入到 fw
		if _, err = io.Copy(fw, tr); err != nil {
			return
		}

		if err = fw.Chmod(info.Mode().Perm()); err != nil {
			return
		}

		slog.Info(fmt.Sprintf("  - %s => %s ...ok", h.Name, dstPath))
		return
	}

	slog.Info("extract embed xunlei", "target", target)
	tr := tar.NewReader(d)
	for h, e := tr.Next(); e != io.EOF; h, e = tr.Next() {
		if e != nil {
			err = e
		}

		if err != nil {
			break
		}

		if h.Typeflag == tar.TypeDir {
			continue
		}

		err = extract(tr, h)
	}

	return
}
