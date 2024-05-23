package xlp

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	SYNOPKG_PKGNAME           = "pan-xunlei-com"                               //包名
	SYNOPKG_PKGDEST           = "/var/packages/" + SYNOPKG_PKGNAME + "/target" //包安装目录
	SYNOPKG_DSM_VERSION_MAJOR = "7"                                            //系统的主版本
	SYNOPKG_DSM_VERSION_MINOR = "2"                                            //系统的次版本
	SYNOPKG_DSM_VERSION_BUILD = "64570"                                        //系统的编译版本

	PAN_XUNLEI_CLI = SYNOPKG_PKGDEST + "/bin/bin/xunlei-pan-cli-launcher." + runtime.GOARCH //主进程名称
	PID_FILE       = SYNOPKG_PKGDEST + "/var/" + SYNOPKG_PKGNAME + ".pid"                   //进程文件

	LAUNCHER_LISTEN_PATH = SYNOPKG_PKGDEST + "/var/pan-xunlei-com-launcher.sock" //启动器监听地址
	DRIVE_LISTEN_PATH    = SYNOPKG_PKGDEST + "/var/pan-xunlei-com.sock"          //主程序监听地址

	ROOT_PATH     = "/var/packages/" + SYNOPKG_PKGNAME + "/shares/迅雷" //主目录
	CONFIG_PATH   = ROOT_PATH + "/"                                   //
	DOWNLOAD_PATH = ROOT_PATH + "/下载"                                 //默认下载目录
	DRIVE_PATH    = ROOT_PATH + "/.drive"                             //数据目录
	LOG_PAN       = DRIVE_PATH + "/log-pan.log"                       //日志文件
	LOG_CGI       = DRIVE_PATH + "/log-cgi.log"                       //日志文件
	LOG_LAUNCHER  = DRIVE_PATH + "/log-launcher.log"                  //日志文件

	PATH_SYNO_INFO_CONF          = "/etc/synoinfo.conf"                                            //synoinfo.conf 文件路径
	PATH_SYNO_AUTHENTICATE_CGI   = "/usr/syno/synoman/webman/modules/authenticate.cgi"             //syno...authenticate.cgi 文件路径
	SYNO_AUTHENTICATE_CGI_SCRIPT = "#!/bin/sh\necho Content-Type: text/plain\necho;\necho admin\n" //syno...authenticate.cgi 文件内容
	UPDATE_URL                   = "/webman/3rdparty/" + SYNOPKG_PKGNAME + "/version"
)

var (
	SYNOPKG_PKGVER = cat(filepath.Join(SYNOPKG_PKGDEST, "bin/bin/version"))                                                                 //包版本
	SYNOPLATFORM   = Iif(runtime.GOARCH == "amd64", "apollolake", "rtd1296")                                                                //平台
	OS_VERSION     = SYNOPLATFORM + " dsm " + SYNOPKG_DSM_VERSION_MAJOR + "." + SYNOPKG_DSM_VERSION_MINOR + "-" + SYNOPKG_DSM_VERSION_BUILD //系统版本
)

// 默认配置，端口2345，下载保存文件夹 /downloads, 数据文件夹 /data, 关闭调试日志
func New() *Daemon {
	d := &Daemon{}
	d.cfg.Init()
	return d
}

// 模拟环境启动器
type Daemon struct {
	cfg Config
	ver string
}

func (d *Daemon) Version(ver string) *Daemon {
	d.ver = ver
	return d
}

func (d *Daemon) BindFlag(fs flagSet, parse bool) *Daemon {
	d.cfg.BindFlag(fs, parse)
	return d
}

func (d *Daemon) IsDebug() bool { return d.cfg.Debug }

func (d *Daemon) Run(ctx context.Context) {
	if d.cfg.DirDownload != "" {
		d.cfg.DirDownload, _ = filepath.Abs(d.cfg.DirDownload)
	}

	if d.cfg.DirData != "" {
		d.cfg.DirData, _ = filepath.Abs(d.cfg.DirData)
	}

	d.cfg.DashboardPort = SelectN(d.cfg.DashboardPort, 2345)
	d.cfg.DirDownload = SelectN(d.cfg.DirDownload, "/xunlei/downloads")
	d.cfg.DirData = SelectN(d.cfg.DirData, "/xunlei/data")

	err := func() (err error) {
		if err = GroupCall(ctx, createDir(d.cfg.DirData), createDir(d.cfg.DirDownload)); err != nil {
			return
		}
		mode := d.runMode()
		slog.Info("run", "mode", mode)
		switch mode {
		case "fork":
			err = d.forkRun(ctx)
		case "chroot":
			err = d.chrootRun(ctx)
		default:
			err = d.directRun(ctx)
		}
		return
	}()

	if err != nil {
		slog.Error("exited!", "err", fmt.Sprintf("%+v", err))
	} else {
		slog.Info("exited!")
	}
}

func (d *Daemon) runMode() string {
	switch {
	case d.cfg.Chroot != "" && d.isDirect():
		return "fork"
	case !d.isDirect():
		return "chroot"
	default:
		return "direct"
	}
}

func (d *Daemon) directRun(ctx context.Context) (err error) {
	return d.run(ctx)
}

func (d *Daemon) forkRun(ctx context.Context) (err error) {
	return NewRoot(d.cfg.Chroot).AppendOSEnv(d.cfg.MarshalEnv()...).
		Bind("/dev", "/lib", "/etc", "/var", "/bin", "/tmp", "/usr").
		BindOptional("/lib32", "/libx32", "/lib64", "/run", "/root", "/mnt").
		Bind(d.cfg.DirData, d.cfg.DirDownload).
		Run(ctx)
}

func (d *Daemon) chrootRun(ctx context.Context) (err error) {
	if err = sysChroot(d.getRoot()); err != nil {
		err = fmt.Errorf("change root: %w", err)
		return
	}
	slog.Debug("root changed", "root", d.getRoot())
	err = d.run(ctx)
	return
}

// 启动
func (d *Daemon) run(ctx context.Context) (err error) {
	if err = checkEnv(); err != nil {
		return
	}

	slog.Info(`_  _ _  _ _  _ _    ____  _`)
	slog.Info(` \/  |  | |\ | |    |___  |`)
	slog.Info(`_/\_ |__| | \| |___ |___  |`)
	slog.Info(fmt.Sprintf(`daemon version: %s`, d.ver))
	slog.Info(fmt.Sprintf(`spk version: v%s`, SYNOPKG_PKGVER))
	slog.Info(``)

	slog.Info("Config")
	slog.Info(fmt.Sprintf("  - DashboardPort:     %s", Iif(d.cfg.DashboardPort == 0, "random", strconv.Itoa(d.cfg.DashboardPort))))
	slog.Info(fmt.Sprintf("  - DashboardUsername: %s", Iif(d.cfg.DashboardUsername == "", "none", d.cfg.DashboardUsername)))
	slog.Info(fmt.Sprintf("  - DashboardPassword: %s", Iif(d.cfg.DashboardPassword == "", "none", d.cfg.DashboardPassword)))
	slog.Info(fmt.Sprintf("  - DownloadPath:      %s", d.cfg.DirDownload))
	slog.Info(fmt.Sprintf("  - DataPath:          %s", d.cfg.DirData))
	slog.Info(fmt.Sprintf("  - Debug:             %t", d.cfg.Debug))
	slog.Info(fmt.Sprintf("  - NewRoot:           %s", d.getRoot()))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	env := Env{}.init()
	env.Set("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
	env.WithOS()
	env.Set("SYNOPLATFORM", SYNOPLATFORM)
	env.Set("SYNOPKG_PKGNAME", SYNOPKG_PKGNAME)
	env.Set("SYNOPKG_PKGVER", SYNOPKG_PKGVER)
	env.Set("SYNOPKG_PKGDEST", SYNOPKG_PKGDEST)
	env.Set("SYNOPKG_DSM_VERSION_MAJOR", SYNOPKG_DSM_VERSION_MAJOR)
	env.Set("SYNOPKG_DSM_VERSION_MINOR", SYNOPKG_DSM_VERSION_MINOR)
	env.Set("SYNOPKG_DSM_VERSION_BUILD", SYNOPKG_DSM_VERSION_BUILD)
	env.Set("DriveListen", "unix://"+DRIVE_LISTEN_PATH)
	env.Set("PLATFORM", "群晖")
	env.Set("OS_VERSION", OS_VERSION)
	env.Set("ConfigPath", CONFIG_PATH)
	env.Set("DownloadPATH", DOWNLOAD_PATH)
	env.Set("HOME", DRIVE_PATH)
	env.Set("TLSInsecureSkipVerify", "true")
	env.Set("GIN_MODE", "release")

	slog.Info("Environ")
	env.Each(func(k, v string) { slog.Info(fmt.Sprintf("  - %s=%s", k, v)) })

	var uid, gid uint32

	if uid, gid, err = lookupUg(d.cfg.UID, d.cfg.GID); err != nil {
		err = fmt.Errorf("lookup uid/gid fail: %w", err)
		return
	}

	if err = GroupCall(ctx, createParentDir(DRIVE_LISTEN_PATH)); err != nil {
		return
	}

	// 尝试处理一下权限
	if uid != 0 {
		chown(SYNOPKG_PKGDEST, uid, gid, true)
		chmod(SYNOPKG_PKGDEST, os.ModePerm, true)
		chown(d.cfg.DirData, uid, gid, true)
		chown(d.cfg.DirData, uid, gid, true)
		chown(d.cfg.DirDownload, uid, gid)
		chmod(d.cfg.DirDownload, os.ModePerm)
	}

	if err = GroupCall(ctx,
		symlink(d.cfg.DirData, DRIVE_PATH),
		symlink(d.cfg.DirDownload, DOWNLOAD_PATH),
	); err != nil {
		return
	}

	if err = d.mockSyno(ctx, env.Environ()); err != nil {
		return
	}

	var args = []string{
		"-launcher_listen", "unix://" + LAUNCHER_LISTEN_PATH,
		"-pid", PID_FILE,
		"-logfile", LOG_LAUNCHER,
		"-logsize", "1MB",
	}

	if d.cfg.PreventUpdate {
		args = append(args, "-update_url", fmt.Sprintf("http://127.0.0.1:%d%s", d.cfg.DashboardPort, UPDATE_URL))
	}

	c := exec.CommandContext(ctx, PAN_XUNLEI_CLI, args...)

	setupProcAttr(c, uid, gid)

	c.Dir = SYNOPKG_PKGDEST + "/bin"
	c.Env = env.Environ()

	return loggerRedirect(ctx, c, d.IsDebug())
}

func (d *Daemon) getRoot() string { return SelectN(os.Getenv(_RUN_WITH_CHROOT), "/") }
func (d *Daemon) isDirect() bool  { return d.getRoot() == "/" }

func (d *Daemon) mockSyno(ctx context.Context, environs []string) (err error) {
	srcPath := filepath.Join(SYNOPKG_PKGDEST, "etc", "synoinfo.conf")

	if err = fileWrite(PATH_SYNO_AUTHENTICATE_CGI, SYNO_AUTHENTICATE_CGI_SCRIPT, os.ModePerm); err != nil {
		return
	}

	if err = fileWriteCopy(srcPath, PATH_SYNO_INFO_CONF, fmt.Sprintf(`unique="synology_%s_720"`, randText(7)), 0); err != nil {
		return
	}

	done := make(chan struct{})

	s := &http.Server{Handler: d.mockHandler(environs), BaseContext: func(net.Listener) context.Context { return ctx }}
	s.Addr = fmt.Sprintf(":%d", d.cfg.DashboardPort)
	slog.Info("syno mock", "addr", s.Addr)

	go func() {
		select {
		case <-done:
			return
		case <-ctx.Done():
			_ = s.Shutdown(context.Background())
		}
	}()

	go func() {
		defer close(done)
		if err := s.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				slog.Error("xlp done!", "err", err)
				return
			}
		}
		slog.Info("xlp is done!")
	}()

	return
}

func (d *Daemon) mockHandler(environs []string) http.Handler {
	redirect := func(to string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, to, http.StatusTemporaryRedirect)
		}
	}

	router := chi.NewMux()
	router.Use(middleware.Recoverer)
	router.Handle("/webman/login.cgi", Respond(fmt.Sprintf(`{"SynoToken":"syno_%s"}`, randText(24)), "application/json", http.StatusOK))
	if d.cfg.PreventUpdate {
		router.Handle("GET "+UPDATE_URL, preventUpdate())
	}

	router.Group(func(r chi.Router) {
		index := fmt.Sprintf("/webman/3rdparty/%s/index.cgi/", SYNOPKG_PKGNAME)
		if d.cfg.DashboardPassword != "" {
			r.Use(middleware.BasicAuth("xlp", d.cfg.UserMap()))
		}

		hcgi := &cgi.Handler{
			Path: fmt.Sprintf("%s/ui/index.cgi", SYNOPKG_PKGDEST),
			Env:  environs,
		}

		r.Handle(index+"*", hcgi)
		r.Handle("GET /", redirect(index))
		r.Handle("GET /web", redirect(index))
		r.Handle("GET /webman", redirect(index))
	})

	return router
}

func chown(path string, uid, gid uint32, r ...bool) {
	walkFiles(path, func(cur string) {
		if err := os.Chown(cur, int(uid), int(gid)); err != nil {
			slog.Warn("chown", "path", cur, "uid", uid, "gid", gid, "err", err)
		} else {
			slog.Debug("chown", "path", cur, "uid", uid, "gid", gid)
		}
	}, len(r) > 0 && r[0])
}

func chmod(path string, perm fs.FileMode, r ...bool) {
	walkFiles(path, func(cur string) {
		if err := os.Chmod(cur, perm); err != nil {
			slog.Warn("chmod", "path", cur, "perm", perm.String(), "err", err)
		} else {
			slog.Debug("chmod", "path", cur, "perm", perm.String())
		}
	}, len(r) > 0 && r[0])
}

func walkFiles(root string, do func(cur string), r bool) {
	if r {
		filepath.WalkDir(root, func(cur string, _ fs.DirEntry, _ error) (err error) { do(cur); return })
	} else {
		do(root)
	}
}

func preventUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Respond(fmt.Sprintf("arch: %s\nversion: \"0.0.1\"\naccept: [\"9.9.9\"]", runtime.GOARCH), `text/vnd.yaml`, 200)
	}
}
