package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	SYNOPKG_DSM_VERSION_MAJOR = "7"
	SYNOPKG_DSM_VERSION_MINOR = "0"
	SYNOPKG_DSM_VERSION_BUILD = "1"
	SYNOPKG_PKGNAME           = "pan-xunlei-com"
	SYNOPKG_PKGBASE           = "/var/packages/" + SYNOPKG_PKGNAME
	SYNOPKG_PKGDEST           = SYNOPKG_PKGBASE + "/target"
	SYNOPKG_VAR               = SYNOPKG_PKGDEST + "/var/"
	LAUNCHER_EXE              = SYNOPKG_PKGDEST + "/xunlei-pan-cli-launcher"
	LAUNCHER_SOCK             = "unix://" + SYNOPKG_VAR + SYNOPKG_PKGNAME + "-launcher.sock"
	SOCK_FILE                 = "unix://" + SYNOPKG_VAR + SYNOPKG_PKGNAME + ".sock"
	PID_FILE                  = SYNOPKG_VAR + SYNOPKG_PKGNAME + ".pid"
	ENV_FILE                  = SYNOPKG_VAR + SYNOPKG_PKGNAME + ".env"
	LOG_FILE                  = SYNOPKG_VAR + SYNOPKG_PKGNAME + ".log"
	LAUNCH_PID_FILE           = SYNOPKG_VAR + SYNOPKG_PKGNAME + "-launcher.pid"
	LAUNCH_LOG_FILE           = SYNOPKG_VAR + SYNOPKG_PKGNAME + "-launcher.log"
	INST_LOG                  = SYNOPKG_VAR + SYNOPKG_PKGNAME + "_install.log"

	CONFIG_PATH = SYNOPKG_PKGBASE + "/shares/"
	HOME_PATH   = CONFIG_PATH
)

type XunleiDaemon struct {
	Port        int    `json:"port"`
	Internal    bool   `json:"internal"`
	DownloadDIR string `json:"dir"`
	closers     []func(ctx context.Context) error
}

func (d *XunleiDaemon) Run(ctx context.Context, args []string) error {
	defer d.Start().Stop()
	<-ctx.Done()
	return nil
}

func (d *XunleiDaemon) Usage() string {
	return "运行"
}

func (d *XunleiDaemon) Start() *XunleiDaemon {
	log := Standard("启动")
	if err := d.loadConfig(); err != nil {
		log.Fatalf("加载配置文件出错: %v", err)
	}

	go func() {
		if err := d.startEngine(); err != nil {
			log.Warnf("%v", err)
		}
	}()

	go func() {
		if err := d.startUI(); err != nil {
			log.Warnf("%v", err)
		}
	}()

	return d
}

func (d *XunleiDaemon) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for _, closer := range d.closers {
		_ = closer(ctx)
	}
}

func (d *XunleiDaemon) loadConfig() error {
	data, err := os.ReadFile(filepath.Join(SYNOPKG_PKGBASE, "config.json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, d)
}

func (d *XunleiDaemon) address() string {
	if d.Internal {
		return fmt.Sprintf("127.0.0.1:%d", d.Port)
	}
	return fmt.Sprintf(":%d", d.Port)
}

func (d *XunleiDaemon) getEnv() (environs []string) {
	environs = os.Environ()
	// environs = append(environs, `PLATFORM="`+d.Name+`"`)
	environs = append(environs, `DriveListen=`+SOCK_FILE)
	environs = append(environs, fmt.Sprintf(`OS_VERSION="dsm %s.%s-%s"`, SYNOPKG_DSM_VERSION_MAJOR, SYNOPKG_DSM_VERSION_MINOR, SYNOPKG_DSM_VERSION_BUILD))
	environs = append(environs, `HOME=`+HOME_PATH)
	environs = append(environs, `ConfigPath=`+CONFIG_PATH)
	environs = append(environs, `DownloadPATH=`+d.DownloadDIR)
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
	environs = append(environs, "INST_LOG="+INST_LOG)
	environs = append(environs, "GIN_MODE=release")
	return
}

// 启动面板
func (d *XunleiDaemon) startUI() error {
	log := Standard("面板")
	mux := chi.NewMux()
	mux.Use(middleware.Recoverer)

	home := "/webman/3rdparty/" + SYNOPKG_PKGNAME + "/index.cgi"
	// 跳转
	jump := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) { http.Redirect(rw, r, home+"/", 307) })
	mux.Handle("/", jump)
	mux.Handle("/webman/", jump)
	mux.Handle("/webman/3rdparty/"+SYNOPKG_PKGNAME, jump)

	// 迅雷面板CGI
	mux.Handle(home+"/*",
		&cgi.Handler{
			Path: filepath.Join(SYNOPKG_PKGDEST, "xunlei-pan-cli-web"),
			Root: SYNOPKG_PKGDEST,
			Dir:  SYNOPKG_PKGDEST,
			Env:  d.getEnv(),
		},
	)

	// Mock群晖登录
	mux.Handle("/webman/login.cgi",
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "application/json; charset=utf-8")
			rw.WriteHeader(200)
			rw.Write([]byte(`{"SynoToken":""}`))
		}),
	)

	listenAddress := d.address()
	log.Infof("启动")
	log.Infof("监听端口: %s", listenAddress)

	server := &http.Server{Addr: listenAddress, Handler: mux, BaseContext: func(l net.Listener) context.Context {
		log.Infof("已启动: %s", l.Addr())
		return context.Background()
	}}

	d.closers = append(d.closers, server.Shutdown)
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Infof("已退出")
		} else {
			log.Warnf("已退出: %v", err)
			return err
		}
	}
	return nil
}

// 启动服务
func (d *XunleiDaemon) startEngine() error {
	log := Standard("下载引擎")
	os.MkdirAll(SYNOPKG_VAR, 0755)
	daemon := exec.Command(LAUNCHER_EXE, "-launcher_listen="+LAUNCHER_SOCK, "-pid="+PID_FILE, "-logfile="+LAUNCH_LOG_FILE)
	daemon.Dir = SYNOPKG_PKGDEST
	daemon.Env = d.getEnv()
	daemon.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	daemon.Stderr = os.Stderr
	daemon.Stdout = os.Stdout
	daemon.Stdin = os.Stdin

	log.Infof("启动")
	log.Infof("命令行: %s", daemon.String())
	if err := daemon.Start(); err != nil {
		return err
	}

	log.Infof("PID: %d", daemon.Process.Pid)

	d.closers = append(d.closers, func(ctx context.Context) error {
		return syscall.Kill(-daemon.Process.Pid, syscall.SIGINT)
	})

	err := daemon.Wait()
	if daemon.ProcessState != nil {
		log.Infof("已退出: %s", daemon.ProcessState.String())
	} else {
		log.Infof("已退出")
	}
	return err
}
