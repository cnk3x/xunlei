package xlp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
)

// symlink 建立软链接
func symlink(srcPath, dstPath string) func() (err error) {
	return func() (err error) {
		slog.Debug("symlink", "src", srcPath, "dst", dstPath)
		if srcPath, err = filepath.Abs(srcPath); err != nil {
			return
		}

		if dstPath, err = filepath.Abs(dstPath); err != nil {
			return
		}

		if err = createParentDir(dstPath)(); err != nil {
			return
		}

		var dstInfo os.FileInfo
		if dstInfo, err = os.Lstat(dstPath); dstInfo != nil {
			if dstInfo.IsDir() || dstInfo.Mode()&os.ModeSymlink != 0 { //如果是空文件夹或者软链接，删除重建
				err = os.Remove(dstPath)
			} else {
				err = fmt.Errorf("%q is exist, can not make a symlink", dstPath)
			}
		}

		if err != nil && !os.IsNotExist(err) {
			return
		}

		if err = os.Symlink(srcPath, dstPath); err != nil {
			return
		}

		return
	}
}

func createDir(path string) func() error {
	return func() (err error) {
		if info, _ := os.Stat(path); info != nil && info.IsDir() {
			return
		}
		slog.Debug("mkdir", "path", path)
		err = os.MkdirAll(path, os.ModePerm)
		return
	}
}

func createParentDir(path string) func() error {
	return createDir(filepath.Dir(path))
}

// cat 以utf8编码读取文件文本内容
func cat(name string) string {
	d, _ := os.ReadFile(name)
	return string(d)
}

// randText 生成一个指定长度的随机字符串。
func randText(size int) (s string) {
	var d = make([]byte, size/2+1)
	rand.Read(d)
	if s = hex.EncodeToString(d); len(s) > size {
		s = s[:size]
	}
	return
}

func bindEnv(out any, keys ...string) {
	s, found := func(keys ...string) (string, bool) {
		for i, key := range keys {
			if v := os.Getenv(key); v != "" {
				if i > 0 {
					fmt.Printf("[WARN] 环境变量参数%q已过期,请使用%q替代\n", key, keys[0])
				}
				return v, true
			}
		}
		return "", false
	}(keys...)

	if !found {
		return
	}

	switch n := out.(type) {
	case *string:
		*n = s
	default:
		switch n := n.(type) {
		case *bool:
			if r, e := strconv.ParseBool(s); e == nil {
				*n = r
			}
		case *int:
			if r, e := strconv.Atoi(s); e == nil {
				*n = r
			}
		case *int64:
			if r, e := strconv.ParseInt(s, 0, 0); e == nil {
				*n = r
			}
		case *uint:
			if r, e := strconv.ParseUint(s, 0, 0); e == nil {
				*n = uint(r)
			}
		case *uint64:
			if r, e := strconv.ParseUint(s, 0, 0); e == nil {
				*n = r
			}
		case *float64:
			if r, e := strconv.ParseFloat(s, 64); e == nil {
				*n = r
			}
		case *float32:
			if r, e := strconv.ParseFloat(s, 32); e == nil {
				*n = float32(r)
			}
		default:
			panic(fmt.Sprintf("unsupported type: %T: %#v", out, out))
		}
	}
}

// fileWrite 写入文件，存在则跳过
func fileWrite[T ~string | ~[]byte](path string, data T, perm os.FileMode) (err error) {
	slog.Debug("write file", "path", path, "data", string(data))
	if err = createParentDir(path)(); err != nil {
		return
	}

	var f *os.File
	if f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666); err != nil {
		if os.IsExist(err) {
			slog.Warn(fmt.Sprintf("%s has exist, ignore write", path))
			return nil
		}
		return
	}

	_, err = f.Write([]byte(data))

	if err == nil && perm > 0 && perm != 0666 {
		err = f.Chmod(perm)
	}

	if ce := f.Close(); err == nil {
		err = ce
	}

	return
}

// fileWriteCopy 先写入文件到 writeTo，存在则跳过写入，再复制到 copyTo，存在则跳过
func fileWriteCopy[T ~string | ~[]byte](writeTo, copyTo string, data T, perm os.FileMode) (err error) {
	content := []byte(data)
	var stat os.FileInfo
	if stat, err = os.Stat(writeTo); stat != nil && stat.Mode().IsRegular() {
		content, err = os.ReadFile(writeTo)
	} else if os.IsNotExist(err) {
		err = fileWrite(writeTo, content, 0)
	}

	if err != nil {
		return
	}

	return fileWrite(copyTo, content, perm)
}
