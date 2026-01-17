package xunlei

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cgi"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/cnk3x/xunlei/embed/authenticate_cgi"
	"github.com/cnk3x/xunlei/pkg/cmdx"
	"github.com/cnk3x/xunlei/pkg/fo"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/cnk3x/xunlei/pkg/vms/sys"
	"github.com/cnk3x/xunlei/pkg/web"
	"github.com/cnk3x/xunlei/spk"
)

const Version = "3.22.0"

const (
	SYNOPKG_DSM_VERSION_MAJOR = "7"              //系统的主版本
	SYNOPKG_DSM_VERSION_MINOR = "2"              //系统的次版本
	SYNOPKG_DSM_VERSION_BUILD = "64570"          //系统的编译版本
	SYNOPKG_PKGNAME           = "pan-xunlei-com" //包名

	DIR_SYNOPKG_PKGROOT = "/var/packages/" + SYNOPKG_PKGNAME //包安装目录
	DIR_SYNOPKG_PKGDEST = DIR_SYNOPKG_PKGROOT + "/target"    //包安装目录
	DIR_SYNOPKG_WORK    = DIR_SYNOPKG_PKGDEST + "/bin"       //

	FILE_PAN_XUNLEI_VER = DIR_SYNOPKG_PKGDEST + "/bin/bin/version"                                   //版本文件
	FILE_PAN_XUNLEI_CLI = DIR_SYNOPKG_PKGDEST + "/bin/bin/xunlei-pan-cli-launcher." + runtime.GOARCH //启动器
	FILE_INDEX_CGI      = DIR_SYNOPKG_PKGDEST + "/ui/index.cgi"                                      //CGI文件路径

	DIR_VAR              = DIR_SYNOPKG_PKGDEST + "/var"                       //SYNOPKG_PKGROOT
	FILE_PID             = DIR_VAR + "/" + SYNOPKG_PKGNAME + ".pid"           //进程文件
	SOCK_LAUNCHER_LISTEN = DIR_VAR + "/" + SYNOPKG_PKGNAME + "-launcher.sock" //启动器监听地址
	SOCK_DRIVE_LISTEN    = DIR_VAR + "/" + SYNOPKG_PKGNAME + ".sock"          //主程序监听地址

	FILE_SYNO_INFO_CONF        = "/etc/synoinfo.conf"                                //synoinfo.conf 文件路径
	FILE_SYNO_AUTHENTICATE_CGI = "/usr/syno/synoman/webman/modules/authenticate.cgi" //syno...authenticate.cgi 文件路径
)

var (
	SYNO_PLATFORM = utils.Iif(runtime.GOARCH == "amd64", "geminilake", "rtd1296")                                                           //平台
	SYNO_MODEL    = utils.Iif(runtime.GOARCH == "amd64", "DS920+", "DS220j")                                                                //平台
	SYNO_VERSION  = SYNO_PLATFORM + " dsm " + SYNOPKG_DSM_VERSION_MAJOR + "." + SYNOPKG_DSM_VERSION_MINOR + "-" + SYNOPKG_DSM_VERSION_BUILD //系统版本
)

func Before(cfg Config) func(ctx context.Context) (func(), error) {
	return func(ctx context.Context) (undo func(), err error) {
		u := utils.MakeUndoPool(&undo, &err)
		defer u.ErrDefer()

		undo, err = mockSyno(ctx, cfg.Chroot)
		if err != nil {
			return
		}
		u.Put(undo)

		target := filepath.Join(cfg.Chroot, DIR_SYNOPKG_PKGDEST)
		varPath := filepath.Join(cfg.Chroot, DIR_VAR)
		dirHome := filepath.Join(cfg.DirData, ".drive")

		undo, err = sys.Mkdirs(ctx, append(cfg.DirDownload, target, varPath, dirHome), 0777)
		if err != nil {
			return
		}
		u.Put(undo)

		dest := filepath.Join(cfg.Chroot, DIR_SYNOPKG_PKGDEST)
		if err = spk.Download(ctx, cfg.SpkUrl, dest, cfg.ForceDownload); err != nil {
			return
		}

		if uid, gid := cfg.Uid, cfg.Gid; uid != 0 || gid != 0 {
			gid = cmp.Or(gid, uid)
			sys.Chown(ctx, utils.Array(varPath, dirHome), uid, gid, false)
			sys.Chown(ctx, utils.Array(cfg.DirData, target), uid, gid, true)
		}
		return
	}
}

func Run(cfg Config) func(ctx context.Context) error {
	return func(ctx context.Context) (err error) {
		slog.InfoContext(ctx, "app start")
		defer slog.InfoContext(ctx, "app done")

		ctx, cancel := context.WithCancelCause(ctx)
		defer cancel(fmt.Errorf("done"))

		var dirDownload []string
		if dirDownload, err = utils.NewRootPath(cfg.Chroot, cfg.DirDownload...); err != nil {
			return
		}

		var dirData []string
		if dirData, err = utils.NewRootPath(cfg.Chroot, cfg.DirData); err != nil {
			return
		}

		envs := mockEnv(dirData[0], strings.Join(dirDownload, ":"))
		return cmdx.Exec(
			log.Prefix(ctx, "vms"),
			FILE_PAN_XUNLEI_CLI,
			cmdx.Flags(
				"-launcher_listen", "unix://"+SOCK_LAUNCHER_LISTEN,
				"-pid", FILE_PID,
				"-update_url", utils.Iif(cfg.PreventUpdate, "null", ""),
				"-logfile", cfg.LauncherLogFile,
			),
			cmdx.Dir(DIR_SYNOPKG_WORK),
			cmdx.Env(envs),
			cmdx.LineErr(logPan("pan", "stderr")),
			cmdx.LineOut(logPan("pan", "stdout")),
			cmdx.OnStarted(func(c *cmdx.Cmd) error {
				utils.BackExec(func() {
					if err := webRun(ctx, envs, cfg); err != nil {
						cancel(err)
					} else {
						cancel(fmt.Errorf("web done"))
					}
				})
				return nil
			}),
		)
	}
}

func webRun(ctx context.Context, env []string, cfg Config) (err error) {
	ctx = log.Prefix(ctx, "web")
	mux := web.NewMux()
	mux.Recoverer()

	c1 := cmdx.LineWriter(logPan("pan", "cgi"))
	defer c1.Close()

	c2 := cmdx.LineWriter(func() func(s string) {
		r := regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})\s+`)
		return func(s string) {
			var t slog.Attr
			if matches := r.FindStringSubmatch(s); len(matches) > 0 {
				if d, e := time.Parse("2026/01/16 22:08:13", matches[1]); e == nil {
					t = slog.Time(slog.TimeKey, d)
				} else {
					t = slog.String("pan_time", matches[1])
				}
				s = s[len(matches[0]):]
			}
			slog.LogAttrs(ctx, slog.LevelDebug, "[cgi] [err] "+s, log.PrefixAttr("pan"), t)
		}
	}())
	defer c2.Close()

	mux.Handle("/webman/status",
		web.FBlob(
			func() (string, error) { return "hello xlp, " + time.Now().Format(time.RFC3339), nil },
			"text/plain",
		),
	)

	const CGI_PATH = "/webman/3rdparty/" + SYNOPKG_PKGNAME + "/index.cgi/" //

	mux.With(web.BasicAuth(cfg.DashboardUsername, cfg.DashboardPassword)).Mount(CGI_PATH, &cgi.Handler{
		Dir:    DIR_SYNOPKG_WORK,
		Path:   FILE_INDEX_CGI,
		Env:    env,
		Logger: utils.LogStd(c2),
		Stderr: c1,
	})

	mux.Get("/", web.Redirect(CGI_PATH, true))
	mux.Get("/web", web.Redirect(CGI_PATH, true))
	mux.Get("/webman", web.Redirect(CGI_PATH, true))
	mux.Handle("/webman/login.cgi", web.Blob(fmt.Sprintf(`{"SynoToken":%q,"result":"success","success":true}`, utils.RandText(13)), "application/json", http.StatusOK))

	// mux.Get("/webman/3rdparty/"+SYNOPKG_PKGNAME+"/version", web.Blob(fmt.Sprintf("arch: %s\nversion: \"0.0.1\"\naccept: [\"9.9.9\"]", runtime.GOARCH), `text/vnd.yaml`, 200))
	return mux.Run(ctx, web.Address(cfg.Ip, cfg.Port))
}

func mockEnv(dirData, dirDownload string) []string {
	ld_lib := os.Getenv("LD_LIBRARY_PATH")
	return append(os.Environ(),
		"SYNOPLATFORM="+SYNO_PLATFORM,
		"SYNOPKG_PKGNAME="+SYNOPKG_PKGNAME,
		"SYNOPKG_PKGDEST="+DIR_SYNOPKG_PKGDEST,
		"SYNOPKG_DSM_VERSION_MAJOR="+SYNOPKG_DSM_VERSION_MAJOR,
		"SYNOPKG_DSM_VERSION_MINOR="+SYNOPKG_DSM_VERSION_MINOR,
		"SYNOPKG_DSM_VERSION_BUILD="+SYNOPKG_DSM_VERSION_BUILD,
		"DriveListen=unix://"+SOCK_DRIVE_LISTEN,
		"PLATFORM=群晖",
		"OS_VERSION="+SYNO_VERSION,
		"ConfigPath="+dirData,
		"HOME="+filepath.Join(dirData, ".drive"),
		"DownloadPATH="+dirDownload,
		"GIN_MODE=release",
		"LD_LIBRARY_PATH=/lib"+utils.Iif(ld_lib == "", "", ":")+ld_lib,
	)
}

func mockSyno(ctx context.Context, root string) (undo func(), err error) {
	return utils.SeqExecWithUndo(
		mockSynoInfo(ctx, root),
		authenticate_cgi.SaveFunc(ctx, filepath.Join(root, FILE_SYNO_AUTHENTICATE_CGI)),
	)
}

func mockSynoInfo(ctx context.Context, root string) func() (undo func(), err error) {
	return func() (undo func(), err error) {
		return fo.WriteFile(ctx,
			filepath.Join(root, FILE_SYNO_INFO_CONF),
			fo.Lines(
				fmt.Sprintf(`platform_name=%q`, SYNO_PLATFORM),
				fmt.Sprintf(`synobios=%q`, SYNO_PLATFORM),
				fmt.Sprintf(`unique=synology_%s_%s`, SYNO_PLATFORM, SYNO_MODEL),
			),
		)
	}
}

func logPan(module, prefix string) func(string) {
	var l slog.Level
	p := log.PrefixAttr(module)
	// 2026-01-17T12:51:53.176432382+08:00
	// 2026-01-16T22:08:13.883171573+08:00 INFO|WARNING|ERROR
	r := regexp.MustCompile(`([0-9T:.+-]+)\s+(INFO|ERROR|WARNING)\s*>?\s*`)
	if prefix != "" {
		prefix = fmt.Sprintf("[%s] ", prefix)
	}

	timeParse := func(s string) (time.Time, bool) {
		if t, e := time.Parse("2006-01-02T15:04:05.999999999-0700", s); e == nil {
			return t, true
		}
		t, e := time.Parse(time.RFC3339Nano, s)
		if e == nil {
			return t, true
		}
		return time.Time{}, false
	}

	return func(s string) {
		var t slog.Attr
		if matches := r.FindStringSubmatch(s); len(matches) > 0 {
			l = log.LevelFromString(matches[2], slog.LevelDebug)
			if d, ok := timeParse(matches[1]); ok {
				t = slog.Time(slog.TimeKey, d)
			} else {
				t = slog.String("pan_time", matches[1])
			}
			s = s[len(matches[0]):]
		}
		slog.LogAttrs(context.Background(), l, prefix+s, p, t)
	}
}
