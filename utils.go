package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

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

func downloadFile(uri string, to string, report func(total, cur, bps float64)) error {
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}

	downloading := to + ".downloading"
	f, err := os.Create(downloading)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(downloading)
	}()

	if report == nil {
		if _, err := io.Copy(f, resp.Body); err != nil {
			return err
		}
	}

	pw := &progress{total: float64(resp.ContentLength), report: report}
	if _, err := io.Copy(io.MultiWriter(f, pw), resp.Body); err != nil {
		return err
	}
	pw.Completed()
	return os.Rename(downloading, to)
}

type progress struct {
	total  float64
	cur    float64
	last   time.Time
	report func(total, cur, bps float64)
}

func (p *progress) Write(b []byte) (n int, err error) {
	n = len(b)
	if p.report != nil {
		p.cur += float64(n)
		if !p.last.IsZero() {
			bps := float64(n) / time.Since(p.last).Seconds()
			go p.report(p.total, p.cur, bps)
		}
		p.last = time.Now()
	}
	return
}

func (p *progress) Completed() {
	if p.report != nil {
		go p.report(p.total, p.total, 0)
	}
}

func getStandardIn() string {
	var value string
	_, err := fmt.Scanln(&value)
	if err == io.EOF { // ctrl+D
		return ""
	}
	if es := err.Error(); es == "unexpected newline" {
		return ""
	}
	return value
}

// 从标准终端读取一行数据，ctrl+D 取消
func scanStd(ctx context.Context, tip string, validate func() error) (string, error) {
	var value string
	for {
		fmt.Print(tip)
		value = getStandardIn()

		if _, err := fmt.Scanln(&value); err != nil {
			if err == io.EOF { // ctrl+D
				return "", err
			}
			if ce := ctx.Err(); ce != nil {
				return "", io.EOF
			}

			// 直接回车了
			if es := err.Error(); es == "unexpected newline" || es == "expected newline" {
				return "", nil
			}
			return "", err
		}
		if err := validate(); err != nil {
			fmt.Printf("[安装]  %v\n", err)
			continue
		}
		break
	}
	return value, nil
}

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

func checkShellPID(pid int, kill bool) error {
	signals := []syscall.Signal{syscall.SIGINT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL}
	sleeps := []time.Duration{time.Second, time.Second * 2, time.Second * 3, time.Second * 3}
	for i := 0; i <= len(signals); i++ {
		if process, _ := os.FindProcess(pid); process != nil {
			if !kill {
				return ErrPIDRunning
			}
			if i == len(signals) {
				return fmt.Errorf("关闭进程[%d]失败", pid)
			}
			if err := syscall.Kill(-pid, signals[i]); err != nil {
				return err
			}
			time.Sleep(sleeps[i])
		}
		break
	}
	return nil
}

var ErrPIDRunning = errors.New("进程在运行")

// https://2rvk4e3gkdnl7u1kl0k.xbase.cloud/v1/file/pancli/versions.info.amd64
//
// func downloadYaml(uri string, out interface{}) error {
// 	resp, err := http.Get(uri)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()
// 	if resp.StatusCode != 200 {
// 		return fmt.Errorf(resp.Status)
// 	}
//
// 	data, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return err
// 	}
// 	return yaml.Unmarshal(data, out)
// }
//
// // walkTar 遍历 tar
// func walkTar(src io.Reader, walkFn func(path string, info fs.FileInfo, body io.Reader) error) error {
// 	tr := tar.NewReader(src)
// 	for {
// 		h, err := tr.Next()
// 		if err != nil {
// 			if err == io.EOF {
// 				return nil
// 			}
// 			return err
// 		}
// 		if err := walkFn(h.Name, h.FileInfo(), tr); err != nil {
// 			if err == fs.SkipDir {
// 				return nil
// 			}
// 			return err
// 		}
// 	}
// }
//
// // txzExtract 解压 tar.xz
// func txzExtract(src io.Reader, dst string) error {
// 	z, err := xz.NewReader(src)
// 	if err != nil {
// 		return err
// 	}
// 	return walkTar(z, func(path string, fi fs.FileInfo, body io.Reader) error {
// 		target := filepath.Join(dst, path)
// 		if fi.Size() > 5<<20 {
// 			infof("  [Extract] %s (%dM)...", target, fi.Size()/(1<<20))
// 		} else {
// 			infof("  [Extract] %s", target)
// 		}
// 		if fi.IsDir() {
// 			return os.MkdirAll(target, fi.Mode())
// 		}
// 		return writeFile(body, target, fi.Mode())
// 	})
// }
//
// // spkExtract 解包SPK
// func spkExtract(spk, target string) error {
// 	src, err := os.Open(spk)
// 	if err != nil {
// 		return err
// 	}
// 	defer src.Close()
//
// 	return walkTar(src, func(path string, _ fs.FileInfo, r io.Reader) error {
// 		if path == "package.tgz" {
// 			if err := txzExtract(r, target); err != nil {
// 				return err
// 			}
// 			return fs.SkipDir
// 		}
// 		return nil
// 	})
// }
