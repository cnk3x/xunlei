package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
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

func xlp(ctx context.Context) (err error) {
	environs := os.Environ()
	environs = append(environs, "SYNOPKG_DSM_VERSION_MAJOR="+SYNOPKG_DSM_VERSION_MAJOR)
	environs = append(environs, "SYNOPKG_DSM_VERSION_MINOR="+SYNOPKG_DSM_VERSION_MINOR)
	environs = append(environs, "SYNOPKG_DSM_VERSION_BUILD="+SYNOPKG_DSM_VERSION_BUILD)
	environs = append(environs, "SYNOPLATFORM="+SYNOPLATFORM)
	environs = append(environs, "SYNOPKG_PKGNAME="+SYNOPKG_PKGNAME)
	environs = append(environs, "SYNOPKG_PKGDEST="+SYNOPKG_PKGDEST)
	environs = append(environs, "HOME="+dataDir)
	environs = append(environs, "DriveListen="+fmt.Sprintf("unix://%s/var/pan-xunlei-com.sock", TARGET_DIR))
	environs = append(environs, "OS_VERSION="+fmt.Sprintf("%s dsm %s.%s-%s", SYNOPLATFORM, SYNOPKG_DSM_VERSION_MAJOR, SYNOPKG_DSM_VERSION_MINOR, SYNOPKG_DSM_VERSION_BUILD))
	environs = append(environs, "DownloadPATH=/迅雷下载")

	if err = os.MkdirAll(filepath.Join(dataDir, "logs"), os.ModePerm); err != nil {
		err = fmt.Errorf("[xlp] 创建日志目录: %w", err)
		return
	}

	if err = os.MkdirAll(filepath.Join(TARGET_DIR, "/var"), os.ModePerm); err != nil {
		err = fmt.Errorf("[xlp] 创建变量目录: %w", err)
		return
	}

	if err = fakeSynoInfo(dataDir); err != nil {
		return
	}

	if err = Chdir(dataDir); err != nil {
		err = fmt.Errorf("[xlp] 跳转到数据库: %w", err)
		return
	}

	c := exec.CommandContext(ctx,
		fmt.Sprintf("%s/bin/xunlei-pan-cli-launcher.%s", TARGET_DIR, runtime.GOARCH),
		"-launcher_listen", fmt.Sprintf("unix://%s/var/pan-xunlei-com-launcher.sock", TARGET_DIR),
		"-pid", fmt.Sprintf("%s/var/pan-xunlei-com-launcher.pid", TARGET_DIR),
	)

	c.Env = environs
	c.SysProcAttr = SetSysProc(&syscall.SysProcAttr{})
	uid, gid := SetUser(c.SysProcAttr)
	if uid > 0 {
		rChown(SYNOPKG_PKGDEST, uid, gid)
		rChown(dataDir, uid, gid)
	}

	if isDebug() {
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		c.Stdin = os.Stdin
	}

	if err = c.Start(); err != nil {
		err = fmt.Errorf("[xlp] [启动器] 启动: %w", err)
		return
	}

	pid := c.Process.Pid
	log.Printf("[xlp] [启动器] 启动成功: %d", pid)

	optPort, _ := strconv.Atoi(os.Getenv(ENV_WEB_PORT))
	if optPort < 0 {
		optPort = 2345
	}
	bindAddr := os.Getenv(ENV_WEB_ADDRESS)
	webPrefix := os.Getenv(ENV_WEB_PREFIX)
	go fakeWeb(ctx, environs, bindAddr, optPort, webPrefix)

	if err = c.Wait(); err != nil {
		err = fmt.Errorf("[xlp] [启动器] 结束: %w", err)
		return
	}

	return
}

func addPrefixRoute(prefix, path string) string {
	if prefix == "" {
		return path
	}
	return filepath.Join(prefix, path)
}
func fakeWeb(ctx context.Context, environs []string, bindAddress string, port int, prefix string) {
	synoToken := []byte(fmt.Sprintf(`{"SynoToken":"syno_%s"}`, randText(24)))
	login := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.WriteHeader(200)
		_, _ = rw.Write(synoToken)
	})
	redirect := func(url string, code int) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			http.Redirect(rw, r, addPrefixRoute(prefix, url), code)
		}
	}

	mux := http.NewServeMux()
	home := fmt.Sprintf("/webman/3rdparty/%s/index.cgi", SYNOPKG_PKGNAME)
	mux.Handle(addPrefixRoute(prefix, "/webman/login.cgi"), login)
	mux.Handle(addPrefixRoute(prefix, "/"), redirect(home+"/", 307))
	mux.Handle(addPrefixRoute(prefix, home), redirect(home+"/", 307))

	indexCGI := &cgi.Handler{Path: addPrefixRoute(prefix, fmt.Sprintf("%s/ui/index.cgi", TARGET_DIR)), Env: environs}
	if !isDebug() {
		indexCGI.Stderr = io.Discard
		indexCGI.Logger = log.New(io.Discard, "", 0)
	} else {
		indexCGI.Stderr = os.Stderr
		indexCGI.Logger = log.Default()
	}

	mux.Handle(addPrefixRoute(prefix, home+"/"), basicAuth(indexCGI))

	s := &http.Server{Addr: fmt.Sprintf("%s:%d", bindAddress, port), Handler: mux}
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

func isDebug() bool {
	debug, _ := strconv.ParseBool(os.Getenv(ENV_DEBUG))
	return debug
}

func randText(size int) string {
	var d = make([]byte, size)
	n, _ := rand.Read(d)
	s := hex.EncodeToString(d[:n])
	if len(s) > size {
		s = s[:size]
	}
	return s
}

func rChown(rootPath string, uid, gid int) {
	if uid == 0 {
		return
	}
	_ = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err := os.Chown(path, uid, gid); err != nil {
			fmt.Fprintf(os.Stderr, "chown: %v", err)
		}
		return nil
	})
}

func basicAuth(next http.Handler) http.Handler {
	if u, p := os.Getenv("XL_BA_USER"), os.Getenv("XL_BA_PASSWORD"); u != "" && p != "" {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || user != u || pass != p {
				w.Header().Add("WWW-Authenticate", `Basic realm="xlp"`)
				w.WriteHeader(401)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
	return next
}
