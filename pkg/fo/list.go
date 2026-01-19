package fo

import (
	"io/fs"
	"path/filepath"
)

func WalkDir(root string, fn func(path string, d fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return fn(path, d)
	})
}
