package xunlei

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/cgi"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/cnk3x/xunlei/pkg/cmd"
	"github.com/cnk3x/xunlei/pkg/flags"
	"github.com/cnk3x/xunlei/pkg/iofs"
	"github.com/cnk3x/xunlei/pkg/lod"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/vms"
	"github.com/cnk3x/xunlei/pkg/web"
	"github.com/cnk3x/xunlei/spk"
)

const (
	SYNOPKG_DSM_VERSION_MAJOR = "7"     //系统的主版本
	SYNOPKG_DSM_VERSION_MINOR = "2"     //系统的次版本
	SYNOPKG_DSM_VERSION_BUILD = "64570" //系统的编译版本

	SYNOPKG_PKGNAME = "pan-xunlei-com"                                     //包名
	SYNOPKG_PKGROOT = "/var/packages/" + SYNOPKG_PKGNAME                   //包安装目录
	SYNOPKG_PKGDEST = SYNOPKG_PKGROOT + "/target"                          //包安装目录
	SYNOPKG_ETC     = SYNOPKG_PKGROOT + "/etc"                             //包安装目录
	PID_FILE        = SYNOPKG_PKGDEST + "/var/" + SYNOPKG_PKGNAME + ".pid" //进程文件

	PAN_XUNLEI_VER = SYNOPKG_PKGDEST + "/bin/bin/version"                                   //版本文件
	PAN_XUNLEI_CLI = SYNOPKG_PKGDEST + "/bin/bin/xunlei-pan-cli-launcher." + runtime.GOARCH //启动器
	PAN_XUNLEI_BIN = SYNOPKG_PKGDEST + "/bin/bin/xunlei-pan-cli.%s." + runtime.GOARCH       //主程序
	PAN_XUNLEI_CGI = SYNOPKG_PKGDEST + "/ui/index.cgi"                                      //CGI文件

	LAUNCHER_LISTEN_PATH = SYNOPKG_PKGDEST + "/var/pan-xunlei-com-launcher.sock" //启动器监听地址
	DRIVE_LISTEN_PATH    = SYNOPKG_PKGDEST + "/var/pan-xunlei-com.sock"          //主程序监听地址

	PATH_SYNO_INFO_CONF        = "/etc/synoinfo.conf"                                //synoinfo.conf 文件路径
	PATH_SYNO_AUTHENTICATE_CGI = "/usr/syno/synoman/webman/modules/authenticate.cgi" //syno...authenticate.cgi 文件路径
	UPDATE_URL                 = "/webman/3rdparty/" + SYNOPKG_PKGNAME + "/version"
)

var (
	SYNO_PLATFORM = lod.Iif(runtime.GOARCH == "amd64", "geminilake", "rtd1296")                                                             //平台
	SYNO_MODEL    = lod.Iif(runtime.GOARCH == "amd64", "DS920+", "DS220j")                                                                  //平台
	OS_VERSION    = SYNO_PLATFORM + " dsm " + SYNOPKG_DSM_VERSION_MAJOR + "." + SYNOPKG_DSM_VERSION_MINOR + "-" + SYNOPKG_DSM_VERSION_BUILD //系统版本
)

const (
	TAG_CHILD  = "child"
	TAG_CHROOT = "chroot"
)

func New(cfg Config, ver string) *Daemon { return &Daemon{cfg: cfg, ver: ver} }

// 模拟环境启动器
type Daemon struct {
	cfg Config
	ver string
}

func (d *Daemon) Run(ctx context.Context) (err error) {
	if err = d.cfg.Validate(); err != nil {
		return
	}

	vm := &vms.Vm{
		Root:  d.cfg.Chroot,
		Binds: append(d.cfg.DirDownload.Abs(), d.cfg.DirData),
		Args:  flags.MustArgs(&d.cfg),
		Env:   os.Environ(),
		Uid:   d.cfg.Uid,
		Gid:   d.cfg.Gid,
		RootRun: func(ctx context.Context) (err error) {
			if iofs.IsFile("/.dockerenv") {
				if err = os.Remove("/.dockerenv"); err != nil {
					return
				}
			}

			auth_cgi_script := "#!/bin/sh\necho Content-Type: text/plain\necho;\necho;\necho admin\n"
			if err = iofs.WriteText(ctx, PATH_SYNO_AUTHENTICATE_CGI, auth_cgi_script, 0777); err != nil {
				return
			}

			infoText := "platform_name=\"%s\"\nsynobios=\"%s\"\nunique=\"synology_%s_%s\""
			infoText = fmt.Sprintf(infoText, SYNO_PLATFORM, SYNO_PLATFORM, SYNO_PLATFORM, SYNO_MODEL)
			err = iofs.WriteText(ctx, PATH_SYNO_INFO_CONF, infoText, 0666)
			return
		},
		UserRun: func(ctx context.Context, uid uint32, gid uint32) (err error) {
			if err = os.MkdirAll(SYNOPKG_PKGROOT, 0777); err != nil {
				return
			}
			if err = os.Chown(SYNOPKG_PKGROOT, int(uid), int(gid)); err != nil {
				return
			}
			return
		},

		Main: d.directRun,
	}

	return vm.Run(ctx)
}

func (d *Daemon) checkEnv(ctx context.Context) (spk_ver string, err error) {
	var repairNeed bool

	if spk_ver = iofs.Cat(PAN_XUNLEI_VER); spk_ver == "" {
		slog.InfoContext(ctx, "can't find package version, repair it")
		repairNeed = true
	}

	if !repairNeed && !iofs.IsExecutable(PAN_XUNLEI_CLI) {
		slog.InfoContext(ctx, "can't find xunlei-pan-cli-launcher, repair it")
		repairNeed = true
	}

	if !repairNeed && !iofs.IsExecutable(PAN_XUNLEI_CGI) {
		slog.InfoContext(ctx, "can't find pan-xunlei-com-cgi, repair it")
		repairNeed = true
	}

	if !repairNeed && !iofs.IsExecutable(fmt.Sprintf(PAN_XUNLEI_BIN, spk_ver)) {
		slog.InfoContext(ctx, "can't find xunlei-pan-cli, repair it", "version", spk_ver)
		repairNeed = true
	}

	if repairNeed {
		if err = spk.ExtractEmbedSpk(ctx, SYNOPKG_PKGDEST); err != nil {
			return
		}
	}

	if spk_ver = iofs.Cat(PAN_XUNLEI_VER); spk_ver == "" {
		err = fmt.Errorf("can't find package version")
		return
	}

	return
}

// 启动
func (d *Daemon) directRun(ctx context.Context) (err error) {
	var spk_ver string
	if spk_ver, err = d.checkEnv(ctx); err != nil {
		return
	}

	slog.InfoContext(ctx, `_  _ _  _ _  _ _    ____  _`)
	slog.InfoContext(ctx, ` \/  |  | |\ | |    |___  |`)
	slog.InfoContext(ctx, `_/\_ |__| | \| |___ |___  |`)
	slog.InfoContext(ctx, fmt.Sprintf(`%s %s`, "daemon version:", d.ver))
	slog.InfoContext(ctx, fmt.Sprintf(`%s %s`, "spk version:", spk_ver))
	slog.InfoContext(ctx, fmt.Sprintf("%s %d", "port:", d.cfg.Port))
	slog.InfoContext(ctx, fmt.Sprintf("%s %s", "ip:", d.cfg.Ip))
	slog.InfoContext(ctx, fmt.Sprintf("%s %s", "dashboard username:", d.cfg.DashboardUsername))
	slog.InfoContext(ctx, fmt.Sprintf("%s %s", "dashboard password:", d.cfg.DashboardPassword))
	slog.InfoContext(ctx, fmt.Sprintf("%s %s", "dir download:", d.cfg.DirDownload))
	slog.InfoContext(ctx, fmt.Sprintf("%s %s", "dir data:", d.cfg.DirData))
	slog.InfoContext(ctx, fmt.Sprintf("%s %s", "uid:", d.cfg.Uid))
	slog.InfoContext(ctx, fmt.Sprintf("%s %s", "gid:", d.cfg.Gid))
	slog.InfoContext(ctx, fmt.Sprintf("%s %t", "prevent update:", d.cfg.PreventUpdate))
	slog.InfoContext(ctx, fmt.Sprintf("%s %s", "chroot:", d.cfg.Chroot))
	slog.InfoContext(ctx, fmt.Sprintf("%s %t", "debug:", d.cfg.Debug))
	cfgArgs, _ := flags.Args(&d.cfg)
	slog.InfoContext(ctx, fmt.Sprintf(`%s "%s"`, "args:", strings.Join(cfgArgs, `" "`)))

	if _, err = iofs.Mkdir(filepath.Dir(DRIVE_LISTEN_PATH), 0777); err != nil {
		return
	}

	env := cmd.EnvSet(os.Environ()).Clean().
		Set("SYNOPLATFORM", SYNO_PLATFORM).
		Set("SYNOPKG_PKGNAME", SYNOPKG_PKGNAME).
		Set("SYNOPKG_PKGVER", spk_ver).
		Set("SYNOPKG_PKGDEST", SYNOPKG_PKGDEST).
		Set("SYNOPKG_DSM_VERSION_MAJOR", SYNOPKG_DSM_VERSION_MAJOR).
		Set("SYNOPKG_DSM_VERSION_MINOR", SYNOPKG_DSM_VERSION_MINOR).
		Set("SYNOPKG_DSM_VERSION_BUILD", SYNOPKG_DSM_VERSION_BUILD).
		Set("DriveListen", "unix://"+DRIVE_LISTEN_PATH).
		Set("PLATFORM", "群晖").
		Set("OS_VERSION", OS_VERSION).
		Set("ConfigPath", d.cfg.DirData).
		Set("HOME", filepath.Join(d.cfg.DirData, ".drive")).
		Set("DownloadPATH", d.cfg.DirDownload.String()).
		Set("TLSInsecureSkipVerify", "true").
		Set("GIN_MODE", "release")

	var port uint16
	if port, err = d.mockCgi(ctx, env); err != nil {
		return
	}

	args := []string{"-launcher_listen", "unix://" + LAUNCHER_LISTEN_PATH, "-pid", PID_FILE}
	if d.cfg.PreventUpdate {
		args = append(args, "-update_url", fmt.Sprintf("http://127.0.0.1:%d%s", port, UPDATE_URL))
	}

	err = cmd.Run(ctx, PAN_XUNLEI_CLI, args, cmd.RunOption{
		Dir: SYNOPKG_PKGDEST + "/bin",
		Env: env,
		Tag: "xunlei-pan-cli",
		Log: handleXlLog(),
	})

	return
}

func (d *Daemon) mockCgi(ctx context.Context, environs []string) (port uint16, err error) {
	ctx = log.Prefix(ctx, "mock")

	mux := web.NewMux()

	mux.Group(func(r web.Router) {
		r.Handle("/webman/login.cgi", web.Blob(fmt.Sprintf(`{"SynoToken":%q,"result":"success","success":true}`, iofs.RandText(13)), "application/json", http.StatusOK))
		r.HandleFunc("GET "+UPDATE_URL, func(w http.ResponseWriter, r *http.Request) {
			web.Blob(fmt.Sprintf("arch: %s\nversion: \"0.0.1\"\naccept: [\"9.9.9\"]", runtime.GOARCH), `text/vnd.yaml`, 200)
		})
	})

	wcgi := log.MessageRecive(log.Prefix(ctx, "xunlei-pan-cgi"), handleXlLog())

	mux.Group(func(r web.Router) {
		web.UseBasicAuth(r, d.cfg.DashboardUsername, d.cfg.DashboardPassword)

		hcgi := &cgi.Handler{
			Path:   fmt.Sprintf("%s/ui/index.cgi", SYNOPKG_PKGDEST),
			Env:    environs,
			Stderr: wcgi,
			Logger: log.LogStd(wcgi, "cgi"),
		}

		indexPattern := fmt.Sprintf("/webman/3rdparty/%s/index.cgi/", SYNOPKG_PKGNAME)
		r.Handle(indexPattern+"*", hcgi)
		r.Handle("GET /", web.Redirect(indexPattern, true))
		r.Handle("GET /web", web.Redirect(indexPattern, true))
		r.Handle("GET /webman", web.Redirect(indexPattern, true))
	})

	ip := d.cfg.Ip.String()
	if ip == "<nil>" {
		ip = ""
	}

	return web.Serve(ctx, web.ServeOption{
		Handler: mux,
		Addr:    net.JoinHostPort(ip, strconv.FormatUint(uint64(d.cfg.Port), 10)),
		OnShutDown: func(ctx context.Context) {
			if err := wcgi.Close(); err != nil {
				slog.WarnContext(ctx, "close wcgi", "err", err)
			}
		},
	})
}

func handleXlLog() func(ctx context.Context, s string) {
	var last slog.Level
	locker := sync.Mutex{}

	return func(ctx context.Context, s string) {
		locker.Lock()
		defer locker.Unlock()

		var msg string
		var l = slog.LevelDebug
		var trimFn = func(r rune) bool { return unicode.IsSpace(r) || r == '>' }
		var attrs []slog.Attr

		if strings.HasPrefix(s, `{`) && strings.HasSuffix(s, `}`) {
			msg = s
			var sMap map[string]string
			if err := json.Unmarshal([]byte(s), &sMap); err == nil {
				for k, v := range sMap {
					switch k {
					case "time":
						if t, err := time.Parse(time.RFC3339, v); err != nil {
							attrs = append(attrs, slog.String("parsetime", err.Error()))
						} else {
							attrs = append(attrs, slog.Time(slog.TimeKey, t))
						}
					case "msg":
						if msg, err = url.QueryUnescape(strings.TrimSpace(v)); err != nil {
							msg = v
						}
					case "level":
						l = log.LevelFromString(v)
					default:
						attrs = append(attrs, slog.String(k, v))
					}
				}
			}

			slog.LogAttrs(ctx, l, msg, attrs...)
			return
		}

		switch {
		case strings.Contains(s, "ERROR"):
			_, msg, _ = strings.Cut(s, "ERROR")
			msg = strings.TrimLeftFunc(msg, trimFn)
			l = slog.LevelError
		case strings.Contains(s, "WARNING"):
			_, msg, _ = strings.Cut(s, "WARNING")
			msg = strings.TrimLeftFunc(msg, trimFn)
			l = slog.LevelWarn
		case strings.Contains(s, "WARN"):
			_, msg, _ = strings.Cut(s, "WARN")
			msg = strings.TrimLeftFunc(msg, trimFn)
			l = slog.LevelWarn
		case strings.Contains(s, "INFO"):
			_, msg, _ = strings.Cut(s, "INFO")
			msg = strings.TrimLeftFunc(msg, trimFn)
			l = slog.LevelDebug
		default:
			slog.Log(ctx, last, "- "+s)
			return
		}

		last = l

		t, err := time.ParseInLocation("2006/01/02 15:04:05", strconv.Itoa(time.Now().Year())+"/"+s[:14], time.Local)
		if err != nil {
			attrs = append(attrs, slog.String("parsetime", err.Error()))
		} else {
			attrs = append(attrs, slog.Time(slog.TimeKey, t))
		}

		s = msg
		if msg, err = url.QueryUnescape(strings.TrimSpace(s)); err != nil {
			msg = s
		}

		slog.LogAttrs(ctx, l, msg, attrs...)
	}
}
