package main

import (
	"archive/tar"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xi2/xz"
)

func ExtractXunleiSpk(spk string, dir string) error {
	src, err := os.Open(spk)
	if err != nil {
		return err
	}
	defer src.Close()

	return WalkTar(src, func(header *tar.Header, tr *tar.Reader) (err error) {
		if header.Name == "package.tgz" {
			err = WalkTar(tr, func(header *tar.Header, tr *tar.Reader) (err error) {
				switch {
				case strings.HasPrefix(header.Name, "bin/bin/"):
					err = WriteFileIfNotExist(tr, filepath.Join(dir, "bin", filepath.Base(header.Name)), os.ModePerm)
				case header.Name == "ui/index.cgi":
					err = WriteFileIfNotExist(tr, filepath.Join(dir, "ui", "index.cgi"), os.ModePerm)
				}
				return
			}, xzDecoder)

			if err == nil {
				err = io.EOF
			}
		}
		return
	})
}

func WalkTar(src io.Reader, walkFn func(header *tar.Header, tr *tar.Reader) error, decoders ...func(src io.Reader) (io.ReadCloser, error)) (err error) {
	for _, decoder := range decoders {
		if decoder != nil {
			var rc io.ReadCloser
			if rc, err = decoder(src); err != nil {
				return
			}
			defer rc.Close()
			src = rc
		}
	}

	tr := tar.NewReader(src)
	var h *tar.Header
	for h, err = tr.Next(); err == nil; h, err = tr.Next() {
		if h.FileInfo().IsDir() {
			continue
		}
		err = walkFn(h, tr)
	}

	if err == io.EOF {
		err = nil
	}
	return
}

func WriteFileIfNotExist(r io.Reader, dst string, mode os.FileMode) (err error) {
	log.Printf("write file: %s", dst)

	if err = os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return
	}
	var w *os.File
	if w, err = os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY|os.O_EXCL, mode); err != nil {
		//如果存在就跳过
		if os.IsExist(err) {
			err = nil
			return
		}
		return
	}
	defer w.Close()
	_, err = io.Copy(w, r)
	return
}

func xzDecoder(src io.Reader) (io.ReadCloser, error) {
	zr, err := xz.NewReader(src, 0)
	return io.NopCloser(zr), err
}
