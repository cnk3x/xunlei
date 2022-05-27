package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
)

func syno(ctx context.Context) {
	p, _ := os.Executable()
	p, _ = filepath.EvalSymlinks(p)
	rootfs := filepath.Dir(p)

	binds := []string{
		"/dev", "/sys", "/run", "/lib", "/lib64", "/lib32", "/libx32", "/usr", "/bin",
		"/var/packages/pan-xunlei-com/target",
		"/downloads", "/data",
	}

	files := []string{
		"/etc/resolv.conf",
		"/etc/localtime",
		"/etc/timezone",
		"/etc/hosts",
		"/etc/hosts",
		"/etc/ssl/certs/ca-certificates.crt",
	}

	copies(ctx, rootfs, files...)

	defer umounts(context.Background(), rootfs, binds...)
	umounts(ctx, rootfs, binds...)
	mounts(ctx, rootfs, binds...)

	fatalErr(syscall.Chroot(rootfs))
	fatalErr(syscall.Chdir("/data"))

	xlp(ctx, getOpts())
}

func sh(ctx context.Context, command string, args ...string) error {
	c := exec.CommandContext(ctx, command, args...)
	c.Env = os.Environ()
	c.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
	out, err := c.CombinedOutput()
	if len(out) > 0 {
		lines := bytes.Split(out, []byte("\n"))
		for _, line := range lines {
			if len(line) > 0 {
				log.Println(string(line))
			}
		}
	}
	return err
}

func mounts(ctx context.Context, rootfs string, binds ...string) {
	log.Printf("mounts...")
	for _, src := range binds {
		mount(ctx, src, path.Join(rootfs, src))
	}
}

func mount(ctx context.Context, src, dst string) {
	if i, err := os.Stat(src); err == nil && i.IsDir() {
		if nilErr(os.MkdirAll(dst, 0755), "mount:") {
			nilErr(sh(ctx, "mount", "--bind", src, dst), "mount:")
		}
	}
}

func umounts(ctx context.Context, rootfs string, binds ...string) {
	log.Printf("umounts...")
	for _, src := range binds {
		umount(ctx, path.Join(rootfs, src))
	}
}

func umount(ctx context.Context, path string) {
	if err := sh(ctx, "umount", path); err != nil {
		nilErr(err, "umount:")
	}
}

func copies(ctx context.Context, rootfs string, files ...string) {
	var args []string
	args = append(args, "cp", "--parents")
	args = append(args, files...)
	args = append(args, rootfs)
	if err := sh(ctx, "cp", args...); err != nil {
		nilErr(err, "copies:")
	}
}
