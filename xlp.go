package xunlei

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/cgi"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/cnk3x/xunlei/embed/authenticate_cgi"
	"github.com/cnk3x/xunlei/pkg/cmdx"
	"github.com/cnk3x/xunlei/pkg/fo"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/cnk3x/xunlei/pkg/web"
)

const Version = "3.21"

const (
	SYNOPKG_DSM_VERSION_MAJOR = "7"     //系统的主版本
	SYNOPKG_DSM_VERSION_MINOR = "2"     //系统的次版本
	SYNOPKG_DSM_VERSION_BUILD = "64570" //系统的编译版本

	SYNOPKG_PKGNAME = "pan-xunlei-com"                   //包名
	SYNOPKG_PKGROOT = "/var/packages/" + SYNOPKG_PKGNAME //包安装目录
	SYNOPKG_PKGDEST = SYNOPKG_PKGROOT + "/target"        //包安装目录

	VAR_DIR  = SYNOPKG_PKGDEST + "/var"
	PID_FILE = VAR_DIR + "/" + SYNOPKG_PKGNAME + ".pid" //进程文件

	PAN_XUNLEI_VER = SYNOPKG_PKGDEST + "/bin/bin/version"                                   //版本文件
	PAN_XUNLEI_CLI = SYNOPKG_PKGDEST + "/bin/bin/xunlei-pan-cli-launcher." + runtime.GOARCH //启动器

	LAUNCHER_LISTEN_PATH = VAR_DIR + "/pan-xunlei-com-launcher.sock" //启动器监听地址
	DRIVE_LISTEN_PATH    = VAR_DIR + "/pan-xunlei-com.sock"          //主程序监听地址

	PATH_SYNO_INFO_CONF        = "/etc/synoinfo.conf"                                //synoinfo.conf 文件路径
	PATH_SYNO_AUTHENTICATE_CGI = "/usr/syno/synoman/webman/modules/authenticate.cgi" //syno...authenticate.cgi 文件路径
	UPDATE_URL                 = "/webman/3rdparty/" + SYNOPKG_PKGNAME + "/version"
	CGI_URL                    = "/webman/3rdparty/" + SYNOPKG_PKGNAME + "/index.cgi/"
)

var (
	SYNO_PLATFORM = utils.Iif(runtime.GOARCH == "amd64", "geminilake", "rtd1296")                                                           //平台
	SYNO_MODEL    = utils.Iif(runtime.GOARCH == "amd64", "DS920+", "DS220j")                                                                //平台
	OS_VERSION    = SYNO_PLATFORM + " dsm " + SYNOPKG_DSM_VERSION_MAJOR + "." + SYNOPKG_DSM_VERSION_MINOR + "-" + SYNOPKG_DSM_VERSION_BUILD //系统版本
)

func NewRun(cfg Config) func(ctx context.Context) error {
	return func(ctx context.Context) error { return Run(ctx, cfg) }
}

func NewBefore(cfg Config) func(ctx context.Context) error {
	return func(ctx context.Context) (err error) {
		_ = cmdx.ShellFile(ctx, utils.Eon(os.Executable())+".sh")

		libDir := filepath.Join(cfg.Chroot, "lib")
		for _, dir := range append(cfg.DirDownload, cfg.DirData, filepath.Join(cfg.Chroot, VAR_DIR), libDir) {
			err = os.MkdirAll(dir, 0777)
			slog.Log(ctx, log.ErrDebug(err), "create dir", "dir", dir, "err", err)
			if err != nil {
				return
			}
		}

		path := filepath.Join(cfg.Chroot, PATH_SYNO_AUTHENTICATE_CGI)
		if err = authenticate_cgi.SaveTo(path); os.IsExist(err) {
			err = nil
		}
		slog.Log(ctx, log.ErrDebug(err), "check authenticate.cgi", "path", path, "err", err)
		if err != nil && !os.IsExist(err) {
			return
		}

		err = fo.OpenWrite(
			filepath.Join(cfg.Chroot, PATH_SYNO_INFO_CONF),
			fo.Lines(
				fmt.Sprintf(`platform_name=%q`, SYNO_PLATFORM),
				fmt.Sprintf(`synobios=%q`, SYNO_PLATFORM),
				fmt.Sprintf(`unique=synology_%s_%s`, SYNO_PLATFORM, SYNO_MODEL),
			),
			fo.Perm(0666),
			fo.FlagExcl,
		)

		if errors.Is(err, os.ErrExist) {
			err = nil
		}
		slog.Log(ctx, log.ErrDebug(err), "create file", "path", PATH_SYNO_INFO_CONF, "err", err)
		if err != nil {
			return
		}

		// binDir := filepath.Join(cfg.Chroot, SYNOPKG_PKGDEST, "bin/bin")
		// err = exec.CommandContext(ctx, "sh", "-c", `ldd `+binDir+`/* | grep '=>' | awk '{if ($3 != "") printf "cp %s `+libDir+`/%s\n",$3,$1}' | sh`).Run()
		// slog.Log(ctx, log.ErrDebug(err), "create libs", "path", libDir, "err", err)
		return
	}
}

func Run(ctx context.Context, cfg Config) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer os.RemoveAll(VAR_DIR)

	var dirDownload []string
	if dirDownload, err = utils.NewRootPath(cfg.Chroot, cfg.DirDownload...); err != nil {
		return
	}

	var dirData []string
	if dirData, err = utils.NewRootPath(cfg.Chroot, cfg.DirData); err != nil {
		return
	}

	for _, dir := range append(dirDownload, dirData[0], SYNOPKG_PKGROOT) {
		if err = rChown(ctx, dir, cfg.Uid, cfg.Gid); err != nil {
			return
		}
	}

	envs := mockEnv(dirDownload, dirData[0])

	webDone, e := mockWeb(ctx, envs, cfg)
	if err = e; err != nil {
		return
	}
	utils.After(webDone, cancel)

	slog.DebugContext(ctx, "app start")
	args := []string{"-launcher_listen", "unix://" + LAUNCHER_LISTEN_PATH, "-pid", PID_FILE}

	if cfg.PreventUpdate {
		args = append(args, "-update_url", "null")
	}

	if cfg.LauncherLogFile != "" {
		args = append(args, "-logfile", cfg.LauncherLogFile)
	}

	pan_ctx := log.Prefix(ctx, "pan")
	return cmdx.Exec(
		log.Prefix(ctx, "vms"),
		PAN_XUNLEI_CLI,
		cmdx.Args(args...),
		cmdx.Dir(SYNOPKG_PKGDEST+"/bin"),
		cmdx.Env(envs),
		cmdx.Credential(cfg.Uid, cfg.Gid),
		cmdx.LineRead(func(s string) { slog.DebugContext(pan_ctx, s) }),
	)
}

func mockWeb(ctx context.Context, env []string, cfg Config) (webDone <-chan struct{}, err error) {
	ctx = log.Prefix(ctx, "mock")
	mux := web.NewMux()
	mux.UseRecoverer()

	cgi_ctx := log.Prefix(ctx, "cgi")
	console := cmdx.LineWriter(func(s string) { slog.DebugContext(cgi_ctx, s) })
	hCgi := &cgi.Handler{
		Dir:    fmt.Sprintf("%s/bin", SYNOPKG_PKGDEST),
		Path:   fmt.Sprintf("%s/ui/index.cgi", SYNOPKG_PKGDEST),
		Env:    env,
		Logger: utils.LogStd(console),
		Stderr: console,
	}

	mux.Handle("/", web.Redirect(CGI_URL, true))
	mux.Handle("/web", web.Redirect(CGI_URL, true))
	mux.Handle("/webman", web.Redirect(CGI_URL, true))
	mux.Handle(UPDATE_URL, web.Blob(fmt.Sprintf("arch: %s\nversion: \"0.0.1\"\naccept: [\"9.9.9\"]", runtime.GOARCH), `text/vnd.yaml`, 200))
	mux.Handle("/webman/login.cgi", web.Blob(fmt.Sprintf(`{"SynoToken":%q,"result":"success","success":true}`, utils.RandText(13)), "application/json", http.StatusOK))
	mux.BasicAuth(cfg.DashboardUsername, cfg.DashboardPassword).Route(CGI_URL, hCgi)

	webDone = mux.Start(ctx, web.Address(cfg.Ip, cfg.Port))
	utils.After(webDone, utils.Fne(console.Close))
	return
}

func mockEnv(dirDownload []string, dirData string) []string {
	return append(os.Environ(),
		"SYNOPLATFORM="+SYNO_PLATFORM,
		"SYNOPKG_PKGNAME="+SYNOPKG_PKGNAME,
		"SYNOPKG_PKGDEST="+SYNOPKG_PKGDEST,
		"SYNOPKG_DSM_VERSION_MAJOR="+SYNOPKG_DSM_VERSION_MAJOR,
		"SYNOPKG_DSM_VERSION_MINOR="+SYNOPKG_DSM_VERSION_MINOR,
		"SYNOPKG_DSM_VERSION_BUILD="+SYNOPKG_DSM_VERSION_BUILD,
		"DriveListen=unix://"+DRIVE_LISTEN_PATH,
		"PLATFORM=群晖",
		"OS_VERSION="+OS_VERSION,
		"ConfigPath="+dirData,
		"HOME="+filepath.Join(dirData, ".drive"),
		"DownloadPATH="+strings.Join(dirDownload, string(filepath.ListSeparator)),
		// "TLSInsecureSkipVerify=true",
		"GIN_MODE=release", "LD_LIBRARY_PATH=/lib",
	)
}

func rChown(ctx context.Context, root string, uid, gid uint32) error {
	if uid == 0 {
		return nil
	}

	return filepath.WalkDir(root, func(path string, _ fs.DirEntry, err error) error {
		if err == nil {
			select {
			case <-ctx.Done():
				return fs.SkipAll
			default:
				err = syscall.Chown(path, int(uid), int(gid))
				slog.Log(ctx, log.ErrDebug(err), "chown", "uid", uid, "gid", gid, "path", path, "err", err)
			}
		}
		return err
	})
}

// func readConsole(ctx context.Context) func(string) {
// 	lv := slog.LevelDebug
// 	re0 := regexp.MustCompile(`^\d{2}/\d{2} \d{2}:\d{2}:\d{2}(\.\d+)? (INFO|ERROR|WARNING)\s*>?\s*`)
// 	re1 := regexp.MustCompile(`^[\dTZ:\.-]?\s*(INFO|ERROR|WARNING)\s*(\[\d+\])?\s*`)

// 	return func(s string) {
// 		switch {
// 		case strings.Contains(s, `filter not match`):
// 			return
// 		case strings.Contains(s, `DetectPlatform err:`):
// 			return
// 		case strings.Contains(s, `detect err:key file lost`):
// 			return
// 		case strings.HasPrefix(s, "panic:"):
// 			lv = slog.LevelError
// 		case strings.HasPrefix(s, "RunSafe panic:"):
// 			lv = slog.LevelError
// 		case re0.MatchString(s):
// 			m := re0.FindStringSubmatch(s)
// 			if lv = log.LevelFromString(m[2], slog.LevelDebug); lv == slog.LevelInfo {
// 				lv = slog.LevelDebug
// 			}
// 			s = s[len(m[0]):]
// 		case re1.MatchString(s):
// 			m := re0.FindStringSubmatch(s)
// 			if lv = log.LevelFromString(m[1], slog.LevelDebug); lv == slog.LevelInfo {
// 				lv = slog.LevelDebug
// 			}
// 			s = s[len(m[0]):]
// 		}
// 		s = strings.ReplaceAll(s, `\u0000`, "")
// 		slog.Log(ctx, lv, s)
// 	}
// }

// func panRun(ctx context.Context, cfg Config, env []string) (err error) {
// 	// /data/.drive/bin/xunlei-pan-cli.3.23.5.amd64
// 	// args:[--logsize 10MB --pid /var/packages/pan-xunlei-com/target/var/pan-xunlei-com.pid.child --info /var/packages/pan-xunlei-com/target/bin/bin/info.file -q run runner]
// 	dataRoot := "/" + utils.Eon(filepath.Rel(cfg.Chroot, cfg.DirData))
// 	cliFmt := "xunlei-pan-cli.{version}." + runtime.GOARCH
// 	name, version, find := findVersionBin(filepath.Join(dataRoot, ".drive/bin"), cliFmt)
// 	if !find {
// 		name, version, find = findVersionBin(filepath.Join(SYNOPKG_PKGDEST, "/bin/bin"), cliFmt)
// 		if !find {
// 			return fmt.Errorf("not found xunlei-pan-cli")
// 		}
// 		if err = fo.OpenWrite(filepath.Join(dataRoot, ".drive/bin", ".version"), fo.Content(version), fo.Perm(0666), fo.DirPerm(0777)); err != nil {
// 			return
// 		}
// 		newName := filepath.Join(dataRoot, ".drive/bin", filepath.Base(name))
// 		err = fo.OpenWrite(newName, fo.Content(version), fo.Perm(0666), fo.DirPerm(0777))
// 		if err != nil {
// 			return
// 		}
// 		name = newName
// 	}
// 	args := []string{
// 		"--logsize", "10MB",
// 		"--pid", PID_FILE + ".child",
// 		"--info", filepath.Join(SYNOPKG_PKGDEST, "bin/bin/info.file"),
// 		"-q", "run", "runner",
// 	}
// 	// _ = args
// 	// args = []string{"--help"}
// 	return cmdRun(log.Prefix(ctx, "pan"), name, args, filepath.Join(SYNOPKG_PKGDEST, "bin"), env, cfg.Uid, cfg.Gid)
// }
// func findVersionBin(dir, name string) (bin, version string, find bool) {
// 	version = utils.Cats(filepath.Join(dir, ".version"), filepath.Join(dir, "version"))
// 	if version == "" {
// 		return "", "", false
// 	}
// 	repl := strings.NewReplacer("{version}", version, "{arch}", runtime.GOARCH)
// 	bin = filepath.Join(dir, repl.Replace(name))
// 	find = utils.PathExists(bin)
// 	return
// }
