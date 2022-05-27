package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
)

const (
	SYNOPKG_DSM_VERSION_MAJOR = "7"
	SYNOPKG_DSM_VERSION_MINOR = "1"
	SYNOPKG_DSM_VERSION_BUILD = "0"
	SYNOPLATFORM              = "Synology"
	SYNOPKG_PKGNAME           = "pan-xunlei-com"
	SYNOPKG_PKGDEST           = "/var/packages/" + SYNOPKG_PKGNAME
	TARGET_DIR                = SYNOPKG_PKGDEST + "/target"
)

type Options struct {
	Home         string //数据目录
	DownloadPATH string //下载目录
	Port         int    //网页控制面板访问端口
	Debug        bool   //调试模式，输出迅雷原始的log
}

func getOpts() Options {
	var xlOpts Options
	xlOpts.Home = getEnv("XL_HOME", "/data")
	xlOpts.DownloadPATH = getEnv("XL_DOWNLOAD_PATH", "/downloads")
	xlOpts.Debug, _ = strconv.ParseBool(getEnv("XL_DEBUG", "0"))
	xlOpts.Port, _ = strconv.Atoi(getEnv("XL_WEB_PORT", "2345"))
	return xlOpts
}

func xlp(ctx context.Context, xlOpts Options) {
	log.Printf("[xlp] 数据目录: %s", xlOpts.Home)
	log.Printf("[xlp] 网页端口: %d", xlOpts.Port)
	log.Printf("[xlp] 调试模式: %t", xlOpts.Debug)
	log.Printf("[xlp] 下载目录: %s", xlOpts.DownloadPATH)

	var err error
	xlOpts.Home, err = filepath.Abs(xlOpts.Home)
	fatalErr(err, "[xlp] 数据目录路径错误:")
	xlOpts.DownloadPATH, err = filepath.Abs(xlOpts.DownloadPATH)
	fatalErr(err, "[xlp] 下载目录路径错误:")

	var environs []string
	environs = append(environs, "SYNOPKG_DSM_VERSION_MAJOR="+SYNOPKG_DSM_VERSION_MAJOR)
	environs = append(environs, "SYNOPKG_DSM_VERSION_MINOR="+SYNOPKG_DSM_VERSION_MINOR)
	environs = append(environs, "SYNOPKG_DSM_VERSION_BUILD="+SYNOPKG_DSM_VERSION_BUILD)
	environs = append(environs, "SYNOPLATFORM="+SYNOPLATFORM)
	environs = append(environs, "SYNOPKG_PKGNAME="+SYNOPKG_PKGNAME)
	environs = append(environs, "SYNOPKG_PKGDEST="+SYNOPKG_PKGDEST)
	environs = append(environs, "HOME="+xlOpts.Home)
	environs = append(environs, "DriveListen="+fmt.Sprintf("unix://%s/var/pan-xunlei-com.sock", TARGET_DIR))
	environs = append(environs, "PLATFORM="+SYNOPLATFORM)
	environs = append(environs, "OS_VERSION="+fmt.Sprintf("%s dsm %s.%s-%s", SYNOPLATFORM, SYNOPKG_DSM_VERSION_MAJOR, SYNOPKG_DSM_VERSION_MINOR, SYNOPKG_DSM_VERSION_BUILD))
	environs = append(environs, "DownloadPATH="+xlOpts.DownloadPATH)

	fatalErr(os.MkdirAll(xlOpts.Home, 0777), "[xlp] 创建数据目录:")
	fatalErr(os.MkdirAll(xlOpts.DownloadPATH, 0777), "[xlp] 创建下载目录:")
	fatalErr(os.MkdirAll(xlOpts.Home+"/logs", 0777), "[xlp] 创建日志目录:")
	fatalErr(os.MkdirAll(TARGET_DIR+"/var", 0777), "[xlp] 创建变量目录:")

	fatalErr(fakeSynoInfo(xlOpts.Home))
	fatalErr(os.Chdir(xlOpts.Home), "[xlp] 跳转到数据库:")

	c := exec.CommandContext(ctx,
		fmt.Sprintf("%s/bin/xunlei-pan-cli-launcher.%s", TARGET_DIR, runtime.GOARCH),
		"-launcher_listen", fmt.Sprintf("unix://%s/var/pan-xunlei-com-launcher.sock", TARGET_DIR),
		"-pid", fmt.Sprintf("%s/var/pan-xunlei-com-launcher.pid", TARGET_DIR),
	)

	c.Env = environs
	c.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
	if xlOpts.Debug {
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		c.Stdin = os.Stdin
	}

	fatalErr(c.Start(), "[xlp] [启动器] 启动失败:")
	pid := c.Process.Pid
	log.Printf("[xlp] [启动器] 启动成功: %d", pid)
	go fakeWeb(ctx, environs, xlOpts.Port)
	fatalErr(c.Wait(), "[xlp] [启动器] 已退出:")
}

func fakeWeb(ctx context.Context, environs []string, port int) {
	synoToken := []byte(fmt.Sprintf(`{"SynoToken":"syno_%s"}`, randText(24)))
	login := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.WriteHeader(200)
		_, _ = rw.Write(synoToken)
	})
	redirect := func(url string, code int) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			http.Redirect(rw, r, url, code)
		}
	}

	mux := http.NewServeMux()
	home := fmt.Sprintf("/webman/3rdparty/%s/index.cgi", SYNOPKG_PKGNAME)
	mux.Handle("/webman/login.cgi", login)
	mux.Handle("/", redirect(home+"/", 307))
	mux.Handle(home, redirect(home+"/", 307))
	mux.Handle(home+"/", &cgi.Handler{Path: fmt.Sprintf("%s/ui/index.cgi", TARGET_DIR), Env: environs})

	s := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		<-ctx.Done()
		_ = s.Shutdown(context.Background())
	}()
	log.Printf("[xlp] [UI]启动: %v", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Printf("[xlp] [UI]已退出: %v", err)
			return
		}
	}
	log.Printf("[xlp] [UI]已退出")
}

func fakeSynoInfo(home string) (err error) {
	src := filepath.Join(home, "synoinfo.conf")
	dst := filepath.Join("/etc", "synoinfo.conf")

	if _, err = os.Stat(src); err != nil && os.IsNotExist(err) {
		err = os.WriteFile(src, []byte(fmt.Sprintf(`unique="synology_%s_720"`, randText(7))), 0666)
	}

	if err != nil {
		return fmt.Errorf("[xlp] [复制] %s -> %s: %w", src, dst, err)
	}

	var data []byte
	if data, err = os.ReadFile(src); err != nil {
		return fmt.Errorf("[xlp] [复制] %s -> %s: 读取出错: %w", src, dst, err)
	}

	if err = os.WriteFile(dst, data, 0666); err != nil {
		return fmt.Errorf("[xlp] [复制] %s -> %s: 写入出错: %w", src, dst, err)
	}
	log.Printf("[xlp] [复制] %s -> %s: 成功", src, dst)

	const synoAuthenticate = "/usr/syno/synoman/webman/modules/authenticate.cgi"
	if _, err = os.Stat(synoAuthenticate); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(synoAuthenticate), 0755); err != nil {
			return fmt.Errorf("[xlp] [authenticate.cgi] %s -> %s: 创建目录: %w", src, dst, err)
		}
		if err = os.WriteFile(synoAuthenticate, []byte("#!/bin/sh\necho Content-Type: text/plain\necho\necho admin"), 0755); err != nil {
			return fmt.Errorf("[xlp] [authenticate.cgi] %s -> %s: 写入出错: %w", src, dst, err)
		}
	}

	return
}
