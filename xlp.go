package xunlei

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cgi"
	"os"
	"path/filepath"
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

const Version = "4.0.0"

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

// 还在宿主机环境操作
func Before(cfg Config) func(ctx context.Context) (func(), error) {
	return func(ctx context.Context) (undo func(), err error) {
		bq := utils.BackQueue(&undo, &err)
		defer bq.ErrDefer()

		msUndo, e := mockSyno(ctx, cfg.Root)
		if err = e; err != nil {
			return
		}
		bq.Put(msUndo)

		target := filepath.Join(cfg.Root, DIR_SYNOPKG_PKGDEST)
		varPath := filepath.Join(cfg.Root, DIR_VAR)
		dirHome := filepath.Join(cfg.DirData, ".drive")

		if e := os.RemoveAll(varPath); e != nil {
			slog.WarnContext(ctx, "clean dir fail", "path", varPath, "err", e)
		}

		mdUndo, e := sys.Mkdirs(ctx, append(cfg.DirDownload, target, varPath, dirHome), 0777)
		if err = e; err != nil {
			return
		}
		bq.Put(mdUndo)

		dest := filepath.Join(cfg.Root, DIR_SYNOPKG_PKGDEST)
		if err = spk.Download(ctx, cfg.SpkUrl, dest, cfg.SpkForceDownload); err != nil {
			return
		}

		if cfg.Uid != 0 || cfg.Gid != 0 {
			sys.Chown(ctx, utils.Array(varPath), cfg.Uid, cfg.Gid, false)
			sys.Chown(ctx, utils.Array(cfg.DirData, target), cfg.Uid, cfg.Gid, true)
		}

		return
	}
}

// 已chroot的操作
func Run(cfg Config) func(ctx context.Context) error {
	return func(ctx context.Context) (err error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		fixPath := func(s string) (rel string, err error) {
			if rel, err = filepath.Rel(cfg.Root, s); err != nil {
				return "", err
			}
			if strings.HasPrefix(rel, ".") {
				return "", fmt.Errorf("%s is not a subpath of %s", rel, cfg.Root)
			}
			return "/" + rel, nil
		}

		dirDownload := cfg.DirDownload
		if dirDownload, err = utils.Replace(dirDownload, fixPath); err != nil {
			return
		}

		dirData := cfg.DirData
		if dirData, err = fixPath(cfg.DirData); err != nil {
			return
		}

		envs := mockEnv(dirData, strings.Join(dirDownload, ":"))

		go webRun(log.Prefix(ctx, "web"), envs, cfg, cancel)

		panCtx := log.Prefix(ctx, "pan")
		return cmdx.Run(ctx, FILE_PAN_XUNLEI_CLI,
			cmdx.Flags(
				"-launcher_listen", "unix://"+SOCK_LAUNCHER_LISTEN,
				"-pid", FILE_PID,
				"-update_url", utils.Iif(cfg.PreventUpdate, "null", ""),
				"-logfile", cfg.LauncherLogFile,
			),
			cmdx.Dir(DIR_SYNOPKG_WORK),
			cmdx.Env(envs),
			cmdx.LineErr(func(s string) { slog.DebugContext(panCtx, "[err] "+s) }),
			cmdx.LineOut(func(s string) { slog.DebugContext(panCtx, "[std] "+s) }),
			cmdx.PreStart(func(c *cmdx.Cmd) error {
				ctx := log.Prefix(ctx, "check")
				slog.DebugContext(ctx, "check uid", "uid", os.Geteuid(), "gid", os.Getegid())
				return nil
			}),
			cmdx.OnStarted(func(c *cmdx.Cmd) error {
				slog.InfoContext(ctx, "started", "cmdline", c.String(), "dir", c.Dir, "pid", c.Process.Pid)
				return nil
			}),
			cmdx.OnExit(func(c *cmdx.Cmd) error { cancel(); return cleanExit(DIR_VAR) }),
		)
	}
}

func webRun(ctx context.Context, env []string, cfg Config, onDone func()) {
	defer onDone()
	mux := web.NewMux()
	mux.Recoverer()
	mux.OnStarted(func(addr string) { slog.InfoContext(ctx, "started", "listen", addr) })

	mux.Handle("/webman/status",
		web.FBlob(
			func() (string, error) { return "hello xlp, " + time.Now().Format(time.RFC3339), nil },
			"text/plain",
		),
	)

	lErr := cmdx.LineWriter(func(s string) { slog.DebugContext(ctx, "[cgi] [err] "+s) })
	defer lErr.Close()

	lLog := cmdx.LineWriter(func(s string) { slog.DebugContext(ctx, "[cgi] [log] "+s) })
	defer lLog.Close()

	const CGI_PATH = "/webman/3rdparty/" + SYNOPKG_PKGNAME + "/index.cgi/"
	mux.With(web.BasicAuth(cfg.DashboardUsername, cfg.DashboardPassword)).Mount(CGI_PATH, &cgi.Handler{
		Dir:    DIR_SYNOPKG_WORK,
		Path:   FILE_INDEX_CGI,
		Env:    env,
		Stderr: lErr,
		Logger: utils.LogStd(lLog),
	})

	mux.Get("/", web.Redirect(CGI_PATH, true))
	mux.Get("/web", web.Redirect(CGI_PATH, true))
	mux.Get("/webman", web.Redirect(CGI_PATH, true))
	mux.Handle("/webman/login.cgi", web.Blob(fmt.Sprintf(`{"SynoToken":%q,"result":"success","success":true}`, utils.RandText(13)), "application/json", http.StatusOK))

	err := mux.Run(ctx, web.Address(cfg.Ip, cfg.Port))
	if err != nil {
		slog.WarnContext(ctx, "done", "err", err.Error())
	} else {
		slog.InfoContext(ctx, "done")
	}
}

func mockEnv(dirData, dirDownload string) []string {
	// ld_lib := os.Getenv("LD_LIBRARY_PATH")
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
		// "LD_LIBRARY_PATH=/lib"+utils.Iif(ld_lib == "", "", ":")+ld_lib,
	)
}

func mockSyno(ctx context.Context, root string) (undo func(), err error) {
	bq := utils.BackQueue(&undo, &err)
	defer bq.ErrDefer()

	u1, e := sys.NewFile(ctx, filepath.Join(root, FILE_SYNO_INFO_CONF), fo.Lines(
		fmt.Sprintf(`platform_name=%q`, SYNO_PLATFORM),
		fmt.Sprintf(`synobios=%q`, SYNO_PLATFORM),
		fmt.Sprintf(`unique=synology_%s_%s`, SYNO_PLATFORM, SYNO_MODEL),
	))
	if err = e; err != nil {
		return
	}
	bq.Put(u1)

	u2, e := sys.NewFile(ctx, filepath.Join(root, FILE_SYNO_AUTHENTICATE_CGI), authenticate_cgi.WriteTo)
	if err = e; err != nil {
		return
	}
	bq.Put(u2)

	return
}

func cleanExit(dir string) (err error) {
	err = fo.OpenRead(dir, func(f *os.File) (err error) {
		var entries []os.DirEntry
		if entries, err = f.ReadDir(-1); err != nil {
			return
		}
		for _, entry := range entries {
			if name := entry.Name(); strings.HasSuffix(name, ".sock") || strings.HasSuffix(name, ".pid") {
				if err = os.Remove(filepath.Join(dir, name)); err != nil {
					return
				}
			}
		}
		return
	})

	if os.IsNotExist(err) {
		err = nil
	}

	return
}
