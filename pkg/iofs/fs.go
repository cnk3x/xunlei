package iofs

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/lod"
)

// Cat 以utf8编码读取文件文本内容
func Cat(path string) string {
	return string(ReadFile(path))
}

// WriteText 写入文件
func WriteText[T ~string | ~[]byte](ctx context.Context, path string, data T, perm os.FileMode, overwrite ...bool) (err error) {
	if _, err = Mkdir(filepath.Dir(path), 0777); err != nil {
		return
	}

	err = func() (err error) {
		var f *os.File
		if lod.First(overwrite) {
			f, err = os.Create(path)
		} else {
			f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
		}

		if err != nil {
			if os.IsExist(err) {
				slog.WarnContext(ctx, fmt.Sprintf("%s has exist, ignore write", path))
				return nil
			}
			return
		}

		if len(data) > 0 {
			_, err = f.Write([]byte(data))
		}

		if err == nil && perm > 0 && perm != 0666 {
			err = f.Chmod(perm)
		}

		if ce := f.Close(); err == nil {
			err = ce
		}
		return
	}()

	slog.Log(ctx, lod.ErrDebug(err), "write", "perm", perm, "path", path, "data", string(data), "err", err)
	return
}

// ReadFile reads the named file and returns the contents and ignore the error.
func ReadFile(path string) (data []byte) {
	data, _ = os.ReadFile(path)
	return
}

// CopyFile copies the named file from srcPath to dstPath.
func CopyFile(srcPath, dstPath string, overwrite ...bool) (ok bool, err error) {
	var srcFile *os.File
	if srcFile, err = os.Open(srcPath); err != nil {
		return
	}
	defer srcFile.Close()

	var stat os.FileInfo
	if stat, err = srcFile.Stat(); err != nil {
		return
	}

	if err = os.MkdirAll(filepath.Dir(dstPath), 0777); err != nil {
		return
	}

	var dstFile *os.File
	var flag = os.O_RDWR | os.O_CREATE
	if lod.Select(overwrite...) {
		flag |= os.O_EXCL
	} else {
		flag |= os.O_TRUNC
	}

	if dstFile, err = os.OpenFile(dstPath, flag, stat.Mode().Perm()); err != nil {
		if os.IsExist(err) {
			err = nil
		}
		return
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	ok = err == nil
	return
}

// Readlink 读取软链接信息
func Readlink(linkname string) (dir, name string, err error) {
	if name, err = os.Readlink(linkname); !filepath.IsAbs(name) {
		dir = filepath.Dir(linkname)
	}
	return
}

// Symlink 创建文件系统软链接
//
//	如果目标是软链接，删除
//	如果目标是目录，判断是否为空，为空则删除
//	如果目标是文件，如果overwrite == false报存在错，否则删除
func Symlink(ctx context.Context, oldname, newname string, overwrite ...bool) (err error) {
	defer func() {
		slog.Log(ctx, lod.ErrDebug(err), "symlink", "oldname", oldname, "newname", newname, "err", err)
	}()

	if err = os.MkdirAll(filepath.Dir(newname), 0777); err != nil {
		return
	}

	var stat os.FileInfo
	if stat, err = os.Lstat(newname); err != nil {
		if os.IsNotExist(err) {
			err = os.Symlink(oldname, newname)
		}
		return
	}

	mode := stat.Mode()

	switch {
	case mode&os.ModeSymlink != 0:
		err = os.Remove(newname)
	case mode&os.ModeDir != 0:
		err = os.Remove(newname)
	case mode&os.ModeType == 0:
		if !lod.Select(overwrite...) {
			err = fmt.Errorf("%w: %s", syscall.EISDIR, newname)
		} else {
			err = os.Remove(newname)
		}
	default:
		err = fmt.Errorf("%w: unsupported file type: %s(%s)", os.ErrInvalid, newname, mode.Type().String())
	}

	if err != nil {
		return
	}

	return os.Symlink(oldname, newname)
}

// Mkdir 创建文件夹
//
//	如果文件夹存在，跳过
//	如果目标是软链接，删除，再创建文件夹
//	其他情况，返回目标存在错误
//	返回值 ok 代表新建了文件夹
func Mkdir(dir string, perm os.FileMode) (ok bool, err error) {
	stat, e := os.Lstat(dir)
	if e != nil {
		if os.IsNotExist(e) {
			err = os.MkdirAll(dir, perm)
			ok = err == nil
		}
		return
	}

	switch {
	case stat.IsDir():
		return
	case stat.Mode()&os.ModeSymlink != 0:
		if err = os.Remove(dir); err != nil {
			return
		}
		err = os.MkdirAll(dir, perm)
		ok = err == nil
		return
	default:
		err = fmt.Errorf("%w: %s", os.ErrExist, dir)
	}

	return
}

// IsFile 判断路径是否是文件
func IsFile(path string) bool {
	f, err := os.Stat(path)
	return err == nil && f.Mode().IsRegular()
}

const ModeExecutable fs.FileMode = 0o111

// IsExecutable 判断路径是否可执行
func IsExecutable(path string) bool {
	f, err := os.Stat(path)
	return err == nil && f.Mode().Perm()&ModeExecutable != 0
}

// IsEmptyDir 判断路径是目录且为空
func IsEmptyDir(dir string) (empty bool) {
	stat, err := os.Lstat(dir)
	if err != nil {
		return os.IsNotExist(err) //不存在，当成是空目录
	}

	if !stat.IsDir() {
		return
	}

	d, err := os.Open(dir)
	if err != nil {
		return
	}
	_, err = d.ReadDir(1)
	d.Close()
	return err == io.EOF
}

var B32 = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

// RandText 生成一个指定长度的随机字符串。
func RandText(n int) (s string) {
	var d = make([]byte, (n+4)/8*5)
	_, _ = rand.Read(d)
	if s = B32.EncodeToString(d); len(s) > n {
		s = s[:n]
	}
	return
}
