package utils

import (
	"bytes"
	"cmp"
	"fmt"
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

func PathExists(fn string) bool { return Eol(os.Stat(fn)) == nil }

func Cats(fns ...string) string {
	for _, fn := range fns {
		if s := Cat(fn); s != "" {
			return s
		}
	}
	return ""
}
