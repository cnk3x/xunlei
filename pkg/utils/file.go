package utils

import (
	"bytes"
	"cmp"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// NewRootPath 取相对路径, 并添加`/`前缀, 用于chroot后的路径，如果不是`newRoot`的目录，返回错误
func NewRootPath(newRoot string, targets ...string) (relPath []string, err error) {
	if newRoot, err = filepath.Abs(newRoot); err != nil {
		return
	}

	for _, target := range targets {
		if target, err = filepath.Abs(target); err != nil {
			return
		}

		var r string
		if r, err = filepath.Rel(newRoot, target); err != nil {
			return
		}

		if strings.HasPrefix(r, ".") {
			err = fmt.Errorf("%s is not a subpath of %s", r, newRoot)
			return
		}

		relPath = append(relPath, "/"+r)
	}

	return
}

// Cat 读取文件，输出文件文本内容
//
//	noTrim: 不删除内容前后空白（默认是删除的）
func Cat(fn string, noTrim ...bool) string {
	d, err := os.ReadFile(fn)
	if err != nil {
		return ""
	}
	if !cmp.Or(noTrim...) {
		d = bytes.TrimSpace(d)
	}
	return string(d)
}

// 脱了裤子放个屁，为了能够方便回滚
func Mkdir(dir string, perm fs.FileMode, existsOk bool) (created []string, err error) {
	if dir, err = filepath.Abs(dir); err != nil {
		return
	}

	vol := filepath.VolumeName(dir)
	if vol != "" {
		dir = dir[len(vol):]
	}
	dir = strings.Trim(dir, string(filepath.Separator))

	fields := strings.FieldsFunc(dir, func(r rune) bool { return r == filepath.Separator })
	for i := range fields {
		full := filepath.Join(cmp.Or(vol, string(filepath.Separator)), filepath.Join(fields[:i+1]...))
		switch stat, e := os.Stat(full); {
		case e != nil && !os.IsNotExist(e):
			err = e
		case e != nil:
			if err = os.Mkdir(full, perm); err != nil {
				err = fmt.Errorf("mkdir fail: %w: %s", err, full)
			}
			created = append(created, full)
		case !stat.IsDir():
			err = fmt.Errorf("mkdir fail, parent is not a directory: %w: %s", fs.ErrExist, full)
		case !existsOk:
			err = fmt.Errorf("mkdir fail, parent is exists: %w: %s", fs.ErrExist, full)
		}

		if err != nil {
			return
		}
	}

	return
}
