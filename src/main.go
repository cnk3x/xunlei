package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
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

func main() {
	var (
		root         = "./xunlei"
		home         = "/data"
		downloadPATH = "/downloads"
		port         = 2345
		launcherLog  = ""
		cgiLog       = ""
		httpErrLog   = ""
		arch         = runtime.GOARCH
	)

	flag.StringVar(&root, "root", root, "root directory")
	flag.StringVar(&home, "home", home, "home directory")
	flag.StringVar(&downloadPATH, "download-path", downloadPATH, "download directory")
	flag.IntVar(&port, "port", port, "port")
	flag.StringVar(&launcherLog, "launcher-log", launcherLog, "launcher log type")
	flag.StringVar(&cgiLog, "cgi-log", cgiLog, "cgi log type")
	flag.StringVar(&httpErrLog, "http-err-log", httpErrLog, "http error log type")
	flag.Parse()
	root, _ = filepath.Abs(root)

	log.Printf("root: %s", root)
	log.Printf("home: %s", home)
	log.Printf("port: %d", port)
	log.Printf("launcher-log: %s", launcherLog)
	log.Printf("cgi-log: %s", cgiLog)
	log.Printf("http-err-log: %s", httpErrLog)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	must(fakeSynoInfo(root, home), "fake synoinfo: %v")

	must(syscall.Chroot(root), "chroot: %v")
	must(os.MkdirAll(home+"/logs", 0777), "mkdir logs: %v")
	must(os.MkdirAll(TARGET_DIR+"/var", 0777), "mkdir var: %v")
	must(os.Chdir(home), "chdir: %v")

	must(os.Setenv("SYNOPKG_DSM_VERSION_MAJOR", SYNOPKG_DSM_VERSION_MAJOR))
	must(os.Setenv("SYNOPKG_DSM_VERSION_MINOR", SYNOPKG_DSM_VERSION_MINOR))
	must(os.Setenv("SYNOPKG_DSM_VERSION_BUILD", SYNOPKG_DSM_VERSION_BUILD))
	must(os.Setenv("SYNOPLATFORM", SYNOPLATFORM))
	must(os.Setenv("SYNOPKG_PKGNAME", SYNOPKG_PKGNAME))
	must(os.Setenv("SYNOPKG_PKGDEST", SYNOPKG_PKGDEST))
	must(os.Setenv("HOME", home))

	must(os.Setenv("DriveListen", fmt.Sprintf("unix://%s/var/pan-xunlei-com.sock", TARGET_DIR)))
	must(os.Setenv("PLATFORM", SYNOPLATFORM))
	must(os.Setenv("OS_VERSION", fmt.Sprintf("%s dsm %s.%s-%s", SYNOPLATFORM, SYNOPKG_DSM_VERSION_MAJOR, SYNOPKG_DSM_VERSION_MINOR, SYNOPKG_DSM_VERSION_BUILD)))

	must(os.MkdirAll(downloadPATH, 0777), "mkdir downloads: %v")
	must(os.Setenv("DownloadPATH", downloadPATH))

	c := exec.Command(
		fmt.Sprintf("%s/bin/xunlei-pan-cli-launcher.%s", TARGET_DIR, arch),
		"-launcher_listen", fmt.Sprintf("unix://%s/var/pan-xunlei-com-launcher.sock", TARGET_DIR),
		"-pid", fmt.Sprintf("%s/var/pan-xunlei-com-launcher.pid", TARGET_DIR),
	)
	c.Env = os.Environ()

	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.Dir = home

	logw := createLoggerWriter(launcherLog, home, "launcher")
	defer logw.Close()
	c.Stderr = logw
	c.Stdout = logw
	c.Stdin = os.Stdin

	if err := c.Start(); err != nil {
		log.Fatalf("[启动器]启动失败: %v", err)
	}
	pid := c.Process.Pid
	log.Printf("[启动器]启动成功: %d", pid)

	exited := make(chan struct{})
	defer close(exited)

	go func() {
		select {
		case <-ctx.Done():
			syscall.Kill(-pid, syscall.SIGINT)
		case <-exited:
		}
	}()

	go func() {
		synoToken := []byte(fmt.Sprintf(`{"SynoToken":"syno_%d"}`, randInt(12)))
		login := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "application/json; charset=utf-8")
			rw.WriteHeader(200)
			rw.Write(synoToken)
		})

		cw := createLoggerWriter(cgiLog, home, "cgi")
		defer cw.Close()
		hw := createLoggerWriter(httpErrLog, home, "http")
		defer hw.Close()

		xlWeb := &cgi.Handler{
			Path:   fmt.Sprintf("%s/ui/index.cgi", TARGET_DIR),
			Dir:    SYNOPKG_PKGDEST,
			Env:    os.Environ(),
			Logger: log.New(cw, "[CGI]", log.LstdFlags),
			// Args:   []string{"-dev"},
		}

		mux := http.NewServeMux()
		home := fmt.Sprintf("/webman/3rdparty/%s/index.cgi", SYNOPKG_PKGNAME)
		jump := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) { http.Redirect(rw, r, home+"/", 307) })
		mux.Handle("/", jump)
		mux.Handle("/xunlei", jump)
		mux.Handle(home+"/", xlWeb)
		mux.Handle("/webman/login.cgi", login)
		s := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux, ErrorLog: log.New(hw, "[HTTP]", log.LstdFlags)}
		go func() {
			<-ctx.Done()
			s.Shutdown(context.Background())
		}()
		log.Printf("[UI]启动: %v", s.Addr)
		if err := s.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Printf("[UI]已退出: %v", err)
				return
			}
		}
		log.Printf("[UI]已退出")
	}()

	must(c.Wait(), "启动器已退出: %v")
}

func newRotate(home, name string) io.WriteCloser {
	_ = os.MkdirAll(home+"/logs", 0777)
	return &Rotate{
		Filename:   home + "/logs/" + name + ".log",
		MaxAge:     3,
		MaxBackups: 10,
		LocalTime:  true,
		Compress:   true,
	}
}

func createLoggerWriter(typ string, home, name string) io.WriteCloser {
	var logw io.WriteCloser
	switch typ {
	case "file":
		logw = newRotate(home, name)
	case "stdout", "std":
		logw = nullWriter{os.Stdout}
	case "stderr":
		logw = nullWriter{os.Stderr}
	default:
		logw = nullWriter{}
	}
	return logw
}

func fakeSynoInfo(root, home string) (err error) {
	src := filepath.Join(root, home, "synoinfo.conf")
	dst := filepath.Join(root, "etc", "synoinfo.conf")

	if _, err = os.Stat(src); err != nil && os.IsNotExist(err) {
		err = os.WriteFile(src, []byte(fmt.Sprintf(`unique="synology_%d_720"`, randInt(7))), 0666)
	}

	if err != nil {
		log.Printf("[复制] %s -> %s: %v", src, dst, err)
		return
	}

	var data []byte
	if data, err = os.ReadFile(src); err != nil {
		log.Printf("[复制] %s -> %s: 读取出错: %v", src, dst, err)
		return
	}

	if err = os.WriteFile(dst, data, 0666); err != nil {
		log.Printf("[复制] %s -> %s: 写入出错: %v", src, dst, err)
		return
	}
	log.Printf("[复制] %s -> %s: 成功", src, dst)
	return
}

func randInt(size int) int64 {
	max := int64(math.Pow10(size))
	start := int64(math.Pow10(size - 1))
	return start + rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(max-start)
}

func must(err error, errf ...string) {
	if err != nil {
		if len(errf) > 0 && errf[0] != "" {
			log.Fatalf(errf[0], err)
		} else {
			log.Fatal(err)
		}
	}
}

type nullWriter struct{ io.Writer }

func (n nullWriter) Write(p []byte) (int, error) {
	if n.Writer != nil {
		return n.Writer.Write(p)
	}
	return len(p), nil
}

func (n nullWriter) Close() error {
	return nil
}
