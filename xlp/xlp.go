package xlp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
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
	SYNOPKG_DSM_VERSION_BUILD = "2"                                            //系统的编译版本

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
)

var (
	SYNOPKG_PKGVER = cat(filepath.Join(SYNOPKG_PKGDEST, "bin/bin/version"))                                                                 //包版本
	SYNOPLATFORM   = Iif(runtime.GOARCH == "amd64", "apollolake", "rtd1296")                                                                //平台
	OS_VERSION     = SYNOPLATFORM + " dsm " + SYNOPKG_DSM_VERSION_MAJOR + "." + SYNOPKG_DSM_VERSION_MINOR + "." + SYNOPKG_DSM_VERSION_BUILD //系统版本
)

// 默认配置，端口2345，下载保存文件夹 /downloads, 数据文件夹 /data, 关闭调试日志
func New() *Daemon {
	d := &Daemon{}
	return d
}

// 模拟环境启动器
type Daemon struct{ Config }

func (d *Daemon) WithFlag(fs flagSet) *Daemon {
	d.Config.Flag(fs)
	return d
}

// 启动
func (d *Daemon) Run(ctx context.Context) {
	d.Config.bindEnv()
	d.Config.fill()

	fmt.Printf(`
_  _ _  _ _  _ _    ____  _
 \/  |  | |\ | |    |___  | 
_/\_ |__| | \| |___ |___  | 

version: %s
---------------------------
`, SYNOPKG_PKGVER)

	log.Printf("Config")
	log.Printf("  - Port:             %s", Iif(d.DashboardPort == 0, "random", strconv.Itoa(d.DashboardPort)))
	log.Printf("  - Host:             %s", d.DashboardHost)
	log.Printf("  - User:             %s", Iif(d.DashboardUsername == "", "none", d.DashboardUsername))
	log.Printf("  - Pass:             %s", Iif(d.DashboardPassword == "", "none", d.DashboardPassword))
	log.Printf("  - DownloadPath:     %s", d.DirDownload)
	log.Printf("  - DataPath:         %s", d.DirData)
	log.Printf("  - Logger.Target:    %s", d.Logger.Target)
	log.Printf("  - Logger.Maxsize:   %s", d.Logger.Maxsize)
	log.Printf("  - Logger.Compress:  %s", Iif(d.Logger.Compress, "√", "×"))

	if err := d.run(ctx); err != nil {
		log.Printf("exited: %v", err)
	} else {
		log.Printf("exited!")
	}
}

func (d *Daemon) run(ctx context.Context) (err error) {
	if err = checkEnv(); err != nil {
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	environs := []string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",

		"SYNOPLATFORM=" + SYNOPLATFORM,
		"SYNOPKG_PKGNAME=" + SYNOPKG_PKGNAME,
		"SYNOPKG_PKGVER=" + SYNOPKG_PKGVER,
		"SYNOPKG_PKGDEST=" + SYNOPKG_PKGDEST,
		"SYNOPKG_DSM_VERSION_MAJOR=" + SYNOPKG_DSM_VERSION_MAJOR,
		"SYNOPKG_DSM_VERSION_MINOR=" + SYNOPKG_DSM_VERSION_MINOR,
		"SYNOPKG_DSM_VERSION_BUILD=" + SYNOPKG_DSM_VERSION_BUILD,
		"DriveListen=unix://" + DRIVE_LISTEN_PATH,
		"PLATFORM=群晖",
		"OS_VERSION=" + OS_VERSION,
		"ConfigPath=" + CONFIG_PATH,
		"DownloadPATH=" + DOWNLOAD_PATH,
		"HOME=" + DRIVE_PATH,
		"TLSInsecureSkipVerify=true",
	}

	log.Printf("Environ")
	for _, envIt := range environs {
		log.Printf("  - %s", envIt)
	}

	if err = symlink(d.DirData, DRIVE_PATH); err != nil {
		err = fmt.Errorf("symlink datadir fail: %w", err)
		return
	}

	if err = symlink(d.DirDownload, DOWNLOAD_PATH); err != nil {
		err = fmt.Errorf("symlink download fail: %w", err)
		return
	}

	if err = os.MkdirAll(filepath.Dir(DRIVE_LISTEN_PATH), os.ModePerm); err != nil {
		err = fmt.Errorf("make var dir fail: %w", err)
		return
	}

	if err = d.mockSyno(ctx, environs); err != nil {
		return
	}

	c := exec.CommandContext(
		ctx,
		PAN_XUNLEI_CLI,
		"-launcher_listen", "unix://"+LAUNCHER_LISTEN_PATH,
		"-pid", PID_FILE,
		"-logfile", LOG_LAUNCHER,
		"-logsize", "1MB",
	)

	setupProcAttr(c)

	c.Dir = SYNOPKG_PKGDEST + "/bin"
	c.Env = environs

	w := d.Logger.Create(LOG_PAN)
	defer w.Close()
	c.Stderr = w
	c.Stdout = w

	if err = c.Run(); err != nil {
		return
	}

	return
}

func (d *Daemon) mockSyno(ctx context.Context, environs []string) (err error) {
	//synoinfo: 先保存到应用文件夹内，启动时拷贝到/etc，避免重新部署时机器标识变动
	srcPath := filepath.Join(SYNOPKG_PKGDEST, "etc", "synoinfo.conf")
	if err = fileWriteCopy(srcPath, PATH_SYNO_INFO_CONF, fmt.Sprintf(`unique="synology_%s_720"`, randText(7)), 0); err != nil {
		err = fmt.Errorf("make synoinfo.conf fail: %w", err)
		return
	}

	if err = fileWrite(PATH_SYNO_AUTHENTICATE_CGI, SYNO_AUTHENTICATE_CGI_SCRIPT, os.ModePerm); err != nil {
		err = fmt.Errorf("make syno authenticate.cgi fail: %w", err)
		return
	}

	redirect := func(to string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, to, http.StatusTemporaryRedirect)
		}
	}

	var xllog any

	router := chi.NewMux()
	router.Use(middleware.Recoverer)
	router.Handle("/webman/login.cgi", respond(fmt.Sprintf(`{"SynoToken":"syno_%s"}`, randText(24)), "application/json", http.StatusOK))

	router.Group(func(r chi.Router) {
		index := fmt.Sprintf("/webman/3rdparty/%s/index.cgi/", SYNOPKG_PKGNAME)
		if d.DashboardPassword != "" {
			r.Use(middleware.BasicAuth("xlp", d.UserMap()))
		}

		hcgi := handleCGI(fmt.Sprintf("%s/ui/index.cgi", SYNOPKG_PKGDEST), d.Logger, environs)
		xllog = hcgi.Stderr

		r.Handle(index+"*", hcgi)
		r.Handle("GET /", redirect(index))
	})

	s := &http.Server{Handler: router, BaseContext: func(net.Listener) context.Context { return ctx }}

	done := make(chan struct{})

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
		defer closes(xllog)

		s.Addr = fmt.Sprintf("%s:%d", d.DashboardHost, d.DashboardPort)
		log.Printf("xlp starting, listen at %v", s.Addr)

		if err := s.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Printf("xlp done: %v", err)
				return
			}
		}

		log.Printf("xlp is done")
	}()

	return
}

// symlink 建立软链接
func symlink(srcPath, dstPath string) (err error) {
	if srcPath, err = filepath.Abs(srcPath); err != nil {
		return
	}

	if err = os.MkdirAll(srcPath, os.ModePerm); err != nil {
		return
	}

	if dstPath, err = filepath.Abs(dstPath); err != nil {
		return
	}

	if err = os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		return
	}

	var dstInfo os.FileInfo
	if dstInfo, err = os.Lstat(dstPath); dstInfo != nil {
		if dstInfo.IsDir() || dstInfo.Mode()&os.ModeSymlink != 0 { //如果是空文件夹或者软链接，删除重建
			err = os.Remove(dstPath)
		} else {
			err = fmt.Errorf("%q is exist, can not make a symlink", dstPath)
		}
	}

	if err != nil && !os.IsNotExist(err) {
		return
	}

	if err = os.Symlink(srcPath, dstPath); err != nil {
		return
	}

	return
}

// cat 以utf8编码读取文件文本内容
func cat(name string) string {
	d, _ := os.ReadFile(name)
	return string(d)
}

// fileWrite 写入文件，存在则跳过
func fileWrite[T ~string | ~[]byte](path string, data T, perm os.FileMode) (err error) {
	if err = os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return
	}

	var f *os.File
	if f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666); err != nil {
		if os.IsExist(err) {
			log.Printf("[WARN] %s has exist, ignore write", path)
			return nil
		}
		return
	}

	_, err = f.Write([]byte(data))

	if err == nil && perm > 0 && perm != 0666 {
		err = f.Chmod(perm)
	}

	if ce := f.Close(); err == nil {
		err = ce
	}

	return
}

// fileWrite 先写入文件到 writeTo，存在则跳过写入，再复制到 copyTo，存在则跳过
func fileWriteCopy[T ~string | ~[]byte](writeTo, copyTo string, data T, perm os.FileMode) (err error) {
	content := []byte(data)
	var stat os.FileInfo
	if stat, err = os.Stat(writeTo); stat != nil && stat.Mode().IsRegular() {
		content, err = os.ReadFile(writeTo)
	} else if os.IsNotExist(err) {
		err = fileWrite(writeTo, content, 0)
	}

	if err != nil {
		return
	}

	return fileWrite(copyTo, content, perm)
}

// randText 生成一个指定长度的随机字符串。
func randText(size int) (s string) {
	var d = make([]byte, size/2+1)
	rand.Read(d)
	if s = hex.EncodeToString(d); len(s) > size {
		s = s[:size]
	}
	return
}

// Iif 三元运算
func Iif[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}

func closes(anyCloser any) {
	if anyCloser != nil {
		if closer, ok := anyCloser.(io.Closer); ok {
			closer.Close()
		}
	}
}

func respond[T ~[]byte | ~string](body T, contentType string, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		w.Write([]byte(body))
	}
}

func handleCGI(path string, logger Logger, environs []string) *cgi.Handler {
	l := logger.Create(LOG_CGI)
	return &cgi.Handler{
		Path:   path,
		Stderr: l,
		Env:    environs,
		Logger: log.New(l, "[CGI] ", 0),
	}
}
