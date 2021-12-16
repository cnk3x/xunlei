package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

const (
	SYNOPKG_DSM_VERSION_MAJOR = "7"
	SYNOPKG_DSM_VERSION_MINOR = "0"
	SYNOPKG_DSM_VERSION_BUILD = "1"
	SYNOPKG_PKGNAME           = "pan-xunlei-com"
	SYNOPKG_PKGBASE           = "/var/packages/" + SYNOPKG_PKGNAME
	SYNOPKG_PKGDEST           = SYNOPKG_PKGBASE + "/target"
	LAUNCHER_EXE              = SYNOPKG_PKGDEST + "/bin/bin/xunlei-pan-cli-launcher.amd64"
	LAUNCHER_SOCK             = "unix://" + SYNOPKG_PKGDEST + "/var/" + SYNOPKG_PKGNAME + "-launch.sock"
	LAUNCHER_VAR              = SYNOPKG_PKGDEST + "/var/"
	SOCK_FILE                 = "unix://" + SYNOPKG_PKGDEST + "/var/" + SYNOPKG_PKGNAME + ".sock"
	PID_FILE                  = LAUNCHER_VAR + SYNOPKG_PKGNAME + ".pid"
	ENV_FILE                  = LAUNCHER_VAR + SYNOPKG_PKGNAME + ".env"
	LOG_FILE                  = LAUNCHER_VAR + SYNOPKG_PKGNAME + ".log"
	LAUNCH_PID_FILE           = LAUNCHER_VAR + SYNOPKG_PKGNAME + "-launch.pid"
	LAUNCH_LOG_FILE           = LAUNCHER_VAR + SYNOPKG_PKGNAME + "-launch.log"
	INST_LOG                  = LAUNCHER_VAR + SYNOPKG_PKGNAME + "_install.log"

	CONFIG_PATH = SYNOPKG_PKGBASE + "/shares/"
	HOME_PATH   = CONFIG_PATH + ".drive"
)

const (
	synoAuthenticatePath = "/usr/syno/synoman/webman/modules/authenticate.cgi"
	synoInfoPath         = "/etc/synoinfo.conf"
)

type Options struct {
	Name         string `flag:"name" short:"n" env:"XUNLEI_DEVICE_NAME" usage:"设备名称"`
	Port         int    `flag:"port" short:"p" env:"XUNLEI_PORT" usage:"监听端口"`
	Internal     bool   `flag:"internal" short:"i" env:"XUNLEI_INTERNAL" usage:"仅本机访问"`
	DownloadPATH string `flag:"download-dir" short:"d" env:"XUNLEI_DOWNLOAD_PATH" usage:"下载保存目录"`
}

type App struct {
	Options
	closers []func(ctx context.Context) error
}

func NewApp(options Options) *App {
	return &App{Options: options}
}

func (d *App) preStart() {
	files := []string{
		LAUNCHER_EXE,
		SYNOPKG_PKGDEST + "/bin/bin/xunlei-pan-cli.2.1.0.amd64",
		SYNOPKG_PKGDEST + "/bin/bin/version",
		SYNOPKG_PKGDEST + "/ui/index.cgi",
	}

	ok := true
	log.Println("[启动自检]")
	for _, fn := range files {
		if _, err := os.Stat(fn); err != nil {
			log.Printf("[启动自检] %s: %v", fn, err)
			ok = false
		}
	}
	if !ok {
		log.Fatal("[启动自检] 失败，请查看上面错误信息")
	}

	linkIt := func(p string) error {
		if _, err := os.Stat(p); err != nil {
			if os.IsNotExist(err) {
				os.MkdirAll(filepath.Dir(p), 0755)
				return os.Symlink(filepath.Join(SYNOPKG_PKGBASE, "host", p), p)
			}
			return err
		}
		return nil
	}

	if err := linkIt(synoAuthenticatePath); err != nil {
		log.Fatalf("[启动自检] %v", err)
	}

	if err := linkIt(synoInfoPath); err != nil {
		log.Fatalf("[启动自检] %v", err)
	}
	log.Println("[启动自检] 完成")
}

func (d *App) Start() *App {
	d.preStart()

	go func() {
		if err := d.startEngine(); err != nil {
			log.Printf("%v", err)
		}
	}()

	go func() {
		if err := d.startUI(); err != nil {
			log.Printf("%v", err)
		}
	}()

	return d
}

func (d *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for _, closer := range d.closers {
		if err := closer(ctx); err != nil {
			log.Printf("close: %v", err)
		}
	}
}

func (d *App) address() string {
	if d.Internal {
		return fmt.Sprintf("127.0.0.1:%d", d.Port)
	}
	return fmt.Sprintf(":%d", d.Port)
}

func (d *App) getEnv() (environs []string) {
	environs = os.Environ()
	environs = append(environs, `DriveListen=`+SOCK_FILE)
	environs = append(environs, `PLATFORM="`+d.Name+`"`)
	environs = append(environs, fmt.Sprintf(`OS_VERSION="dsm %s.%s-%s"`, SYNOPKG_DSM_VERSION_MAJOR, SYNOPKG_DSM_VERSION_MINOR, SYNOPKG_DSM_VERSION_BUILD))
	environs = append(environs, `HOME=`+HOME_PATH)
	environs = append(environs, `ConfigPath=`+CONFIG_PATH)
	environs = append(environs, `DownloadPATH=`+d.DownloadPATH)
	environs = append(environs, "SYNOPKG_DSM_VERSION_MAJOR="+SYNOPKG_DSM_VERSION_MAJOR)
	environs = append(environs, "SYNOPKG_DSM_VERSION_MINOR="+SYNOPKG_DSM_VERSION_MINOR)
	environs = append(environs, "SYNOPKG_DSM_VERSION_BUILD="+SYNOPKG_DSM_VERSION_BUILD)
	environs = append(environs, "SYNOPKG_PKGDEST="+SYNOPKG_PKGDEST)
	environs = append(environs, "SYNOPKG_PKGNAME="+SYNOPKG_PKGNAME)
	environs = append(environs, "SVC_CWD="+SYNOPKG_PKGDEST)
	environs = append(environs, "PID_FILE="+PID_FILE)
	environs = append(environs, "ENV_FILE="+ENV_FILE)
	environs = append(environs, "LOG_FILE="+LOG_FILE)
	environs = append(environs, "LAUNCH_LOG_FILE="+LAUNCH_LOG_FILE)
	environs = append(environs, "LAUNCH_PID_FILE="+LAUNCH_PID_FILE)
	environs = append(environs, "INST_LOG="+LAUNCH_PID_FILE)
	environs = append(environs, "GIN_MODE=release")
	return
}

// 启动面板
func (d *App) startUI() error {
	mux := chi.NewMux()
	mux.Use(middleware.Recoverer)
	mux.Use(cors.AllowAll().Handler)
	mux.Use(middleware.Logger)

	// 跳转
	mux.Handle("/",
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			http.Redirect(rw, r, "/webman/3rdparty/"+SYNOPKG_PKGNAME+"/index.cgi/", 307)
		}),
	)

	// 迅雷面板CGI
	mux.Handle("/webman/3rdparty/"+SYNOPKG_PKGNAME+"/index.cgi/*",
		&cgi.Handler{
			Path:       filepath.Join(SYNOPKG_PKGDEST, "ui/index.cgi"),
			Root:       SYNOPKG_PKGDEST,
			Dir:        SYNOPKG_PKGDEST,
			Env:        d.getEnv(),
			InheritEnv: []string{},
			Logger:     &log.Logger{},
			Args:       []string{"--dev"},
			Stderr:     os.Stderr,
		},
	)

	// 迅雷面板资源
	ui := filepath.Join(SYNOPKG_PKGDEST, "ui")
	static := http.FileServer(http.Dir(ui))
	mux.Handle("/webman/3rdparty/"+SYNOPKG_PKGNAME+"/index.cgi/help/*", static)
	mux.Handle("/webman/3rdparty/"+SYNOPKG_PKGNAME+"/index.cgi/images/*", static)
	mux.Handle("/webman/3rdparty/"+SYNOPKG_PKGNAME+"/index.cgi/texts/*", static)

	// Mock群晖登录
	mux.Handle("/webman/login.cgi",
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			render.JSON(rw, r, render.M{"SynoToken": ""}) // 77fa4db9f5dc07f21ff20f5446193dd2
		}),
	)

	listenAddress := d.address()
	log.Printf("[控制面板] 启动")
	log.Printf("[控制面板] 资源目录: %s", ui)
	log.Printf("[控制面板] 监听端口: %s", listenAddress)

	server := &http.Server{Addr: listenAddress, Handler: mux, BaseContext: func(l net.Listener) context.Context {
		log.Printf("[控制面板] 已启动: %s", l.Addr())
		return context.Background()
	}}

	d.closers = append(d.closers, server.Shutdown)
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Printf("[控制面板] 已退出")
		} else {
			log.Printf("[控制面板] 已退出: %v", err)
			return err
		}
	}
	return nil
}

// 启动服务
func (d *App) startEngine() error {
	os.MkdirAll(LAUNCHER_VAR, 0755)
	daemon := exec.Command(LAUNCHER_EXE, "--launcher_listen="+LAUNCHER_SOCK, "--pid="+PID_FILE, "--logfile="+LAUNCH_LOG_FILE)
	daemon.Dir = SYNOPKG_PKGDEST
	daemon.Env = d.getEnv()
	daemon.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	daemon.Stderr = os.Stderr
	daemon.Stdout = os.Stdout
	daemon.Stdin = os.Stdin

	log.Printf("[下载引擎] 启动 ")
	log.Printf("[下载引擎] 命令行: %s", daemon.String())
	if err := daemon.Start(); err != nil {
		return err
	}

	log.Printf("[下载引擎] PID: %d", daemon.Process.Pid)

	d.closers = append(d.closers, func(ctx context.Context) error {
		return syscall.Kill(-daemon.Process.Pid, syscall.SIGINT)
	})

	err := daemon.Wait()
	if daemon.ProcessState != nil {
		log.Printf("[下载引擎] 已退出: %s", daemon.ProcessState.String())
	} else {
		log.Printf("[下载引擎] 已退出")
	}
	return err
}

func delFile(p string) error {
	n, err := os.Readlink(p)
	if err != nil {
		return err
	}
	if strings.HasPrefix(n, SYNOPKG_PKGDEST) {
		if err := os.Remove(p); err != nil {
			return err
		}
		// for {
		// 	p = filepath.Dir(p)
		// 	if yes, err := isEmptyDIR(p); !yes {
		// 		return err
		// 	}
		// 	log.Printf("del dir: %s", p)
		// 	if err := os.Remove(p); err != nil {
		// 		return err
		// 	}
		// }
	}
	return nil
}

func isEmptyDIR(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	names, err := f.Readdirnames(3)
	if err != nil {
		return false, err
	}
	return len(names) == 0, nil
}

type devNull struct{}

func (devNull) Write(p []byte) (n int, err error) {
	return
}

func (devNull) Read(p []byte) (n int, err error) {
	return
}
