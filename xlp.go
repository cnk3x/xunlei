package xunlei

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/cnk3x/xunlei/embed/authenticate_cgi"
	"github.com/cnk3x/xunlei/pkg/fo"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/rootfs"
	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/go-chi/chi/v5"
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
)

var (
	SYNO_PLATFORM = utils.Iif(runtime.GOARCH == "amd64", "geminilake", "rtd1296")                                                           //平台
	SYNO_MODEL    = utils.Iif(runtime.GOARCH == "amd64", "DS920+", "DS220j")                                                                //平台
	OS_VERSION    = SYNO_PLATFORM + " dsm " + SYNOPKG_DSM_VERSION_MAJOR + "." + SYNOPKG_DSM_VERSION_MINOR + "-" + SYNOPKG_DSM_VERSION_BUILD //系统版本
)

func NewRun(cfg Config) func(ctx context.Context) error {
	return func(ctx context.Context) error { return Run(ctx, cfg) }
}

func Run(ctx context.Context, cfg Config) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err = os.MkdirAll(VAR_DIR, 0777); err != nil {
		return
	}

	if err = authenticate_cgi.SaveTo(PATH_SYNO_AUTHENTICATE_CGI); err != nil && !os.IsExist(err) {
		return
	}

	undo3, e := createFile(ctx, PATH_SYNO_INFO_CONF, 0666,
		fmt.Sprintf(`platform_name=%q`, SYNO_PLATFORM),
		fmt.Sprintf(`synobios=%q`, SYNO_PLATFORM),
		fmt.Sprintf(`unique=synology_%s_%s`, SYNO_PLATFORM, SYNO_MODEL),
	)
	if err = e; err != nil {
		return
	}
	defer undo3()

	var dirDownload []string
	if dirDownload, err = utils.NewRootPath(cfg.Chroot, cfg.DirDownload...); err != nil {
		return
	}

	var dirData []string
	if dirData, err = utils.NewRootPath(cfg.Chroot, cfg.DirData); err != nil {
		return
	}

	undo, e := rootfs.Mkdirs(ctx, append(dirDownload, dirData...), 0777, true)
	if err = e; err != nil {
		return
	}
	defer undo()

	for _, dir := range append(dirDownload, dirData[0], SYNOPKG_PKGROOT) {
		if err = rChown(ctx, dir, cfg.Uid, cfg.Gid); err != nil {
			return
		}
	}

	env := makeEnv(dirDownload, dirData[0])

	var port uint16
	var webDone <-chan struct{}
	if port, webDone, err = webRun(ctx, env, cfg); err != nil {
		return
	}

	go utils.SelectDo(webDone, cancel, ctx.Done())

	slog.DebugContext(ctx, "app start")

	args := []string{"-launcher_listen", "unix://" + LAUNCHER_LISTEN_PATH, "-pid", PID_FILE}
	if cfg.PreventUpdate {
		args = append(args, "-update_url", fmt.Sprintf("http://127.0.0.1:%d%s", port, UPDATE_URL))
	}
	return cmdRun(ctx, PAN_XUNLEI_CLI, args, SYNOPKG_PKGDEST+"/bin", env, cfg.Uid, cfg.Gid)
}

func webRun(ctx context.Context, env []string, cfg Config) (port uint16, webDone <-chan struct{}, err error) {
	ctx = log.Prefix(ctx, "mock")
	mux := chi.NewMux()
	mux.Use(webRecoverer)

	console := wrapConsole(ctx)
	hCgi := &cgi.Handler{
		Dir:    fmt.Sprintf("%s/bin", SYNOPKG_PKGDEST),
		Path:   fmt.Sprintf("%s/ui/index.cgi", SYNOPKG_PKGDEST),
		Env:    env,
		Logger: log.Std(console, "wcgi"),
		Stderr: console,
	}

	indexPattern := fmt.Sprintf("/webman/3rdparty/%s/index.cgi/", SYNOPKG_PKGNAME)
	mux.With(webBasicAuth(cfg.DashboardUsername, cfg.DashboardPassword)).Handle(indexPattern+"*", hCgi)

	mux.Handle("GET /", webRedirect(indexPattern, true))
	mux.Handle("GET /web", webRedirect(indexPattern, true))
	mux.Handle("GET /webman", webRedirect(indexPattern, true))
	mux.Handle("/webman/login.cgi", webBlob(fmt.Sprintf(`{"SynoToken":%q,"result":"success","success":true}`, utils.RandText(13)), "application/json", http.StatusOK))
	mux.HandleFunc("GET "+UPDATE_URL, func(w http.ResponseWriter, r *http.Request) {
		webBlob(fmt.Sprintf("arch: %s\nversion: \"0.0.1\"\naccept: [\"9.9.9\"]", runtime.GOARCH), `text/vnd.yaml`, 200)
	})

	ip := cfg.Ip.String()
	if ip == "<nil>" {
		ip = ""
	}

	ports := make(chan uint16, 1)

	err = func() (err error) {
		defer close(ports)
		s := &http.Server{Addr: net.JoinHostPort(ip, strconv.FormatUint(uint64(cfg.Port), 10)), Handler: mux}

		s.BaseContext = func(l net.Listener) context.Context {
			ports <- uint16(l.Addr().(*net.TCPAddr).Port)
			return ctx
		}

		done := make(chan struct{})
		webDone = done

		go func() {
			defer close(done)
			defer console.Close()

			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.ErrorContext(ctx, "ui is done!", "err", err)
			} else {
				slog.InfoContext(ctx, "ui is done!")
			}
		}()

		go utils.SelectDo(ctx.Done(), func() { s.Shutdown(context.Background()) }, done)

		if port, ok, _ := utils.SelectOnce(ports, done); ok {
			slog.InfoContext(ctx, "ui started", "port", port)
		}
		return
	}()
	return
}

func cmdRun(ctx context.Context, name string, args []string, dir string, env []string, uid, gid uint32) (err error) {
	ctx = log.Prefix(ctx, "xunl")
	c := exec.CommandContext(ctx, name, args...)
	c.Dir, c.Env = dir, env
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.Cancel = func() error { return syscall.Kill(-c.Process.Pid, syscall.SIGINT) }
	if uid != 0 || gid != 0 {
		c.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid}
	}

	console := wrapConsole(ctx)
	defer console.Close()
	c.Stdout, c.Stderr = console, console

	if err = c.Start(); err != nil {
		slog.ErrorContext(ctx, "app start fail", "cmd", c.String(), "err", err)
		return
	}
	slog.InfoContext(ctx, "started", "pid", c.Process.Pid, "cmd", c.String())

	err = c.Wait()
	return
}

func wrapConsole(ctx context.Context) io.WriteCloser {
	lv := slog.LevelDebug
	re := regexp.MustCompile(`^\d{2}/\d{2} \d{2}:\d{2}:\d{2}(\.\d+)? (INFO|ERROR|WARNING)\s*>?\s*`)
	r, w := io.Pipe()
	go func() {
		for scan := bufio.NewScanner(r); scan.Scan(); {
			s := scan.Text()
			switch {
			case strings.HasPrefix(s, "panic:"):
				lv = slog.LevelError
			case re.MatchString(s):
				m := re.FindStringSubmatch(s)
				lv = log.LevelFromString(m[2], slog.LevelInfo) - 4
				s = s[len(m[0]):]
			}
			slog.Log(ctx, lv, s)
		}
	}()

	return w
}

func makeEnv(dirDownload []string, dirData string) []string {
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
		"TLSInsecureSkipVerify=true",
		"GIN_MODE=release",
	)
}

func rChown(ctx context.Context, root string, uid, gid uint32) error {
	if uid == 0 {
		return nil
	}

	return filepath.WalkDir(root, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err = context.Cause(ctx); err != nil {
			return err
		}

		slog.DebugContext(ctx, "chown", "uid", uid, "gid", gid, "path", path)
		return os.Chown(path, int(uid), int(gid))
	})
}

func createFile(ctx context.Context, name string, perm fs.FileMode, lines ...string) (undo rootfs.Undo, err error) {
	if undo, err = rootfs.Mkdir(ctx, filepath.Dir(name), 0777, true); err == nil {
		err = fo.OpenWrite(name, fo.Lines(lines...), fo.Perm(perm), fo.FlagExcl)
		if err == nil {
			dirUndo := undo
			undo = func() {
				e := os.Remove(name)
				if e != nil {
					slog.WarnContext(ctx, "remove file", "file", name, "err", e)
				} else {
					slog.DebugContext(ctx, "remove file", "file", name)
				}
				dirUndo()
			}
		}
		if errors.Is(err, os.ErrExist) {
			err = nil
		}
	}

	if err != nil {
		slog.WarnContext(ctx, "create file", "file", name, "err", err)
	} else {
		slog.DebugContext(ctx, "create file", "file", name)
	}
	return
}
