package main

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func DumpFs(fsys fs.FS, name string, target string) error {
	log.Printf("fsys: %s", target)
	src, err := fsys.Open(name)
	if err != nil {
		return err
	}

	info, err := src.Stat()
	if err != nil {
		_ = src.Close()
		return err
	}

	defer os.Chtimes(target, info.ModTime(), info.ModTime())

	if dir := info.IsDir(); !dir {
		return func() (err error) {
			defer src.Close()
			dst, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer dst.Close()
			_, err = io.Copy(dst, src)
			return err
		}()
	}
	src.Close()

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	entities, err := fs.ReadDir(fsys, name)
	if err != nil {
		return err
	}

	for _, entity := range entities {
		n := entity.Name()
		if err := DumpFs(fsys, filepath.Join(name, n), filepath.Join(target, n)); err != nil {
			return err
		}
	}

	return nil
}

func FileCopy(src, dst string) error {
	f1, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f1.Close()

	info, err := f1.Stat()
	if err != nil {
		return err
	}

	f2, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f2.Close()

	_, err = io.Copy(f2, f1)
	if err != nil {
		return err
	}
	return f2.Chmod(info.Mode())
}
