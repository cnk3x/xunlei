package main

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func printIP(format string, log Logger) {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipv4 := ipNet.IP.To4(); ipv4 != nil {
				if !strings.HasPrefix(ipv4.String(), "172") {
					log.Infof(format, ipv4.String())
				}
			} else {
				if !strings.HasPrefix(ipNet.IP.String(), "fe80") {
					log.Infof(format, "["+ipNet.IP.String()+"]")
				}
			}
		}
	}
}

// dumpFs 释放文件
func dumpFs(fsys fs.FS, name string, target string, log Logger) error {
	return fs.WalkDir(fsys, name, func(path string, d fs.DirEntry, err error) error {
		target := filepath.Join(target, strings.TrimPrefix(path, name))
		log.Infof("  [Extract] %s", target)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		src, err := fsys.Open(path)
		if err != nil {
			return err
		}

		return writeFile(src, target, 0755)
	})
}

// CopyFile 复制文件
func copyFile(src, dst string) (err error) {
	var (
		r    *os.File
		info fs.FileInfo
	)

	if r, err = os.Open(src); err != nil {
		return
	}
	defer r.Close()

	if info, err = r.Stat(); err != nil {
		return
	}

	err = writeFile(r, dst, info.Mode())
	return
}

// writeFile 保存为文件
func writeFile(src io.Reader, dst string, mode fs.FileMode) (err error) {
	var w *os.File
	if w, err = os.Create(dst); err != nil {
		return
	}
	defer w.Close()

	if _, err = io.Copy(w, src); err != nil {
		return
	}
	err = w.Chmod(mode)
	return
}

func contextRun(ctx context.Context, runners ...func(context.Context) error) (err error) {
	for _, runner := range runners {
		if err = runner(ctx); err != nil {
			return
		}
	}
	return
}

func serviceControl(args ...string) func(context.Context) error {
	return func(ctx context.Context) error {
		log := Standard("服务")
		cmd := exec.Command("systemctl", args...)
		log.Debugf(cmd.String())
		output, err := cmd.CombinedOutput()
		if output = bytes.TrimSpace(output); len(output) > 0 {
			log.Infof("%s", output)
		} else if err != nil {
			log.Warnf("%v", err)
		}
		return err
	}
}

func serviceControlSilence(args ...string) func(context.Context) error {
	return func(ctx context.Context) error {
		_ = serviceControl(args...)(ctx)
		return nil
	}
}
