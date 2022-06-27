package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
)

func syno(ctx context.Context) (err error) {
	if !isRunInDocker() {
		return fmt.Errorf("[syno] 只能在 docker 中运行")
	}

	optionalBinds := []string{"/run", "/lib", "/lib64", "/lib32", "/libx32", "/usr", "/bin", "/mnt"}
	mustBinds := []string{"/dev", "/sys", TARGET_DIR}

	files := []string{
		"/etc/resolv.conf",
		"/etc/hosts",
		"/etc/localtime",
		"/etc/timezone",
		"/etc/ssl/certs/ca-certificates.crt",
	}

	if _, err = copies(rootfs, files); err != nil {
		return
	}

	var mustBinded []string
	if mustBinded, err = mounts(rootfs, mustBinds, true); err != nil {
		return
	}
	defer umounts(mustBinded...)

	if optionalBinded, _ := mounts(rootfs, optionalBinds); len(optionalBinded) > 0 {
		defer umounts(optionalBinded...)
	}

	p, err := os.Executable()
	if err != nil {
		return err
	}
	c := exec.CommandContext(ctx, p, "daemon")
	c.Stdout = os.Stdout
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr
	c.SysProcAttr = SetSysProc(&syscall.SysProcAttr{})
	c.Env = append(os.Environ(), fmt.Sprintf("%s=%s", ENV_START_AS_SYNO, "1"))

	if err = c.Run(); err != nil {
		err = fmt.Errorf("[syno] down: %w", err)
	}
	return
}

const ENV_START_AS_SYNO = "XL_START_AS_SYNO"

func daemon(ctx context.Context) error {
	if os.Getenv(ENV_START_AS_SYNO) != "1" {
		return fmt.Errorf("[daemon] 请使用 syno 命令启动")
	}

	if err := Chroot(rootfs); err != nil {
		return fmt.Errorf("[daemon] 切换到主目录失败: %w", err)
	}

	if err := Chdir(dataDir); err != nil {
		return fmt.Errorf("[daemon] 切换数据目录失败: %w", err)
	}

	return xlp(ctx)
}

func bind(src, dst string) (err error) {
	stat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("[syno] 源: %w", err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("[syno] 源不是一个目录: %s", src)
	}

	if err = os.MkdirAll(dst, os.ModePerm); err != nil {
		return fmt.Errorf("[syno] 创建目录 %q 失败: %w", dst, err)
	}
	if err = Mount(src, dst); err != nil {
		return fmt.Errorf("[syno] 绑定目录 %q -> %q 失败: %w", src, dst, err)
	}
	return
}

func unbind(path string) (err error) {
	if err = Unmount(path); err != nil {
		return fmt.Errorf("[syno] 卸载目录 %q 失败: %w", path, err)
	}
	return
}

func mounts(rootfs string, binds []string, rollbackOnError ...bool) (binded []string, err error) {
	errRollback := len(rollbackOnError) > 0 && rollbackOnError[0]

	if !errRollback {
		defer func() {
			if err != nil {
				umounts(binded...)
			}
		}()
	}

	for _, src := range binds {
		dst := filepath.Join(rootfs, src)
		if err = bind(src, dst); err != nil {
			if errRollback {
				break
			} else {
				err = nil
				log.Println(err)
			}
		}
		binded = append(binded, dst)
	}
	return
}

func umounts(binds ...string) {
	for _, path := range binds {
		mustUnbind(path)
	}
}

func copies(rootfs string, files []string, overwrite ...bool) (copied []string, err error) {
	defer func() {
		if err != nil {
			for _, f := range copied {
				_ = os.Remove(f)
			}
		}
	}()

	var isCopied bool
	for _, src := range files {
		dst := path.Join(rootfs, src)
		if isCopied, err = fileCopy(src, dst, overwrite...); err != nil {
			err = fmt.Errorf("[syno] 复制文件 %q 到 %q 失败: %v", src, dst, err)
			break
		}
		if isCopied {
			copied = append(copied, dst)
		}
	}
	return
}

func mustUnbind(path string) {
	if err := unbind(path); err != nil {
		log.Printf("%v", err)
	}
}

func fileCopy(src, dst string, overwrite ...bool) (copied bool, err error) {
	var r, w *os.File
	if r, err = os.Open(src); err != nil {
		return
	}
	defer r.Close()

	if err = os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return
	}

	openMode := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	if len(overwrite) > 0 && overwrite[0] {
		openMode |= os.O_EXCL
	}

	if w, err = os.OpenFile(dst, openMode, 0666); err != nil {
		if os.IsExist(err) {
			err = nil
		}
		return
	}

	defer func() {
		if err != nil {
			_ = os.Remove(dst)
		}
	}()

	defer w.Close()

	if _, err = io.Copy(w, r); err != nil {
		return
	}

	var rInfo os.FileInfo
	if rInfo, err = r.Stat(); err != nil {
		return
	}

	if err = w.Chmod(rInfo.Mode()); err != nil {
		return
	}

	copied = true
	return
}

func isRunInDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}
