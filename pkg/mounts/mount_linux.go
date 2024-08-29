package mounts

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/iofs"
	"github.com/cnk3x/xunlei/pkg/lod"
)

// 获取制定目录下的挂载点列表
func MountedEndpoints(dirs ...string) (endpoints []string, err error) {
	var x int
	for _, dir := range dirs {
		if dir != "" {
			if dir, _ = filepath.Abs(dir); dir != "" {
				dirs[x] = dir
				x++
			}
		}
	}

	dirs = dirs[:x]

	hasPrefixes := func(s string) bool {
		if x == 0 {
			return true
		}
		for _, prefix := range dirs {
			if prefix != "" && strings.HasPrefix(s, prefix) {
				return true
			}
		}
		return false
	}

	for br := bufio.NewScanner(bytes.NewReader(iofs.ReadFile("/proc/self/mountinfo"))); br.Scan(); {
		if ls := strings.Fields(br.Text()); len(ls) > 4 {
			if dst := ls[4]; hasPrefixes(dst) {
				endpoints = append(endpoints, dst)
			}
		}
	}

	slices.SortFunc(endpoints, func(a, b string) int { return lod.Select(len(b)-len(a), strings.Compare(a, b)) })

	return
}

func Binds(ctx context.Context, root string, srcPaths []string, continueOnError ...bool) (err error) {
	var errs []error
	slices.SortFunc(srcPaths, func(a, b string) int { return lod.Select(len(a)-len(b), strings.Compare(a, b)) })
	for _, srcPath := range srcPaths {
		if e := Bind(ctx, root, srcPath); e != nil {
			if lod.Selects(continueOnError) {
				errs = append(errs, e)
				continue
			}
			return e
		}
	}

	// return ErrJoin(errs, NewLogWarn(ctx, "bind"))
	return errors.Join(errs...)
}

func Bind(ctx context.Context, root, srcPath string) (err error) {
	defer func() {
		if err == nil {
			slog.DebugContext(ctx, "bind", "rootDir", root, "path", srcPath)
		}
	}()
	if root, err = filepath.Abs(root); err != nil {
		return
	}

	var stat os.FileInfo
	if stat, err = os.Lstat(srcPath); err != nil {
		if os.IsNotExist(err) {
			err = nil
			slog.DebugContext(ctx, "bind - ignored", "path", srcPath, "cause", "lstat - path not exist")
		} else {
			err = fmt.Errorf("bind - lstat: %w, path: %s", err, srcPath)
		}
		return
	}

	mode := stat.Mode()
	target := filepath.Join(root, srcPath)

	switch {
	case mode&os.ModeSymlink != 0: //如果源是软件链接，复制软链接行为，然后向上递归检查并绑定
		var dir, prev string
		if dir, prev, err = iofs.Readlink(srcPath); err != nil {
			return fmt.Errorf("symlink - readlink: %w, path: %s", err, srcPath)
		}

		if err = Bind(ctx, root, filepath.Join(dir, prev)); err != nil {
			return fmt.Errorf("symlink - %w", err)
		}

		if err = iofs.Symlink(ctx, prev, filepath.Join(root, srcPath), true); err != nil {
			return fmt.Errorf("symlink: %w, path: %s", err, prev)
		}
	case mode&os.ModeDir != 0: //如果源目录是文件夹，先判断或创建目标文件夹，再挂载
		if _, err = iofs.Mkdir(target, mode.Perm()); err != nil {
			return fmt.Errorf("dir - mkdirs: %w, target: %s", err, target)
		}
		if err = syscall.Mount(srcPath, target, "auto", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
			return fmt.Errorf("dir - mount: %w", err)
		}
	case mode&os.ModeType == 0: //如果源目录是普通文件，直接复制
		if _, err = iofs.CopyFile(srcPath, target, true); err != nil {
			return fmt.Errorf("file - mount: %w, path: %s, target: %s", err, srcPath, target)
		}
	default:
		return fmt.Errorf("%w: unsupported file type: %s(%s)", os.ErrInvalid, srcPath, mode.Type().String())
	}

	return
}

// 递归卸载挂载点
func Unbind(ctx context.Context, root string, removeEmpty ...bool) (err error) {
	var endpoints []string
	if endpoints, err = MountedEndpoints(root); err != nil {
		return
	}

	var errs []error
	for _, endpoint := range endpoints {
		if e := syscall.Unmount(endpoint, syscall.MNT_DETACH); e != nil {
			errs = append(errs, fmt.Errorf("%w: %s", e, endpoint))
			continue
		}
	}

	for _, endpoint := range endpoints {
		if iofs.IsEmptyDir(endpoint) {
			e := os.Remove(endpoint)
			slog.DebugContext(ctx, "unbind", "endpoint", endpoint, "root", root, "action", "remove dir", "err", e)
		} else {
			slog.DebugContext(ctx, "unbind", "endpoint", endpoint, "root", root)
		}
	}

	// return ErrJoin(errs, NewLogWarn(ctx, "unbind", "root", root))
	return errors.Join(errs...)
}

func Chroot(ctx context.Context, target string) (err error) {
	defer func() { slog.Log(ctx, lod.ErrDebug(err), "chroot", "target", target, "err", err) }()
	if err = os.MkdirAll(target, 0777); err != nil {
		return fmt.Errorf("chroot - mkdir: %w", err)
	}
	if err = os.Chdir(target); err != nil {
		return fmt.Errorf("chroot - chdir: %w", err)
	}
	if err = syscall.Chroot("."); err != nil {
		return fmt.Errorf("chroot: %w", err)
	}
	return
}
