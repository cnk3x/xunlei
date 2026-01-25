package xunlei

import (
	"cmp"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/cnk3x/xunlei/pkg/utils"
)

func Banner(fPrint func(string)) {
	fPrint(`_  _ _  _ _  _ _    ____  _`)
	fPrint(` \/  |  | |\ | |    |___  |`)
	fPrint(`_/\_ |__| | \| |___ |___  |`)
}

// Config 配置
type Config struct {
	//dashboard
	DashboardIp       net.IP `usage:"网页访问绑定IP，默认绑定所有IP" env:"XL_DASHBOARD_IP"`
	DashboardPort     uint16 `usage:"网页访问的端口" env:"XL_DASHBOARD_PORT"`
	DashboardUsername string `usage:"网页访问的用户名" env:"XL_DASHBOARD_USERNAME"` //网页访问的用户名
	DashboardPassword string `usage:"网页访问的密码" env:"XL_DASHBOARD_PASSWORD"`  //网页访问的密码

	//jail 模式
	Root        string   `flag:"r" usage:"主目录，非根目录时“/”需要 root 权限执行" env:"XL_ROOT"`
	DirDownload []string `flag:"d" usage:"下载保存文件夹，可多次指定，需确保有权限访问" env:"XL_DIR_DOWNLOAD"`
	DirData     string   `flag:"D" usage:"程序数据保存文件夹，其下'.drive'文件夹中，存储了登录的账号，下载进度等信息" env:"XL_DIR_DATA"` //程序数据保存文件夹，其下'.drive'文件夹中，存储了登录的账号，下载进度等信息
	Uid         int      `flag:"u" usage:"运行核心的UID，需要 root 权限执行" env:"XL_UID"`
	Gid         int      `flag:"g" usage:"运行核心的GID，需要 root 权限执行" env:"XL_GID"`

	//核心设置
	SpkUrl           string `usage:"下载链接" env:"XL_SPK"`
	SpkForceDownload bool   `usage:"是否强制下载" env:"XL_SPK_FORCE_DOWNLOAD"`
	LauncherLogFile  string `usage:"迅雷启动器日志文件" env:"XL_LAUNCHER_LOG_FILE"`
	PreventUpdate    bool   `usage:"阻止更新" env:"XL_PREVENT_UPDATE"`

	//全局设置
	Debug bool   `usage:"调试模式，调试模式下，失败也不会自动退出，以便于追踪错误，在此模式行，日志等级自动调整为'debug'" env:"XL_DEBUG"`
	Log   string `usage:"日志等级，可选 debug，info，warn，error" env:"XL_LOG"`

	Rootless bool `usage:"无超级用户（root）权限运行，需自行解决权限问题，在此模式下，将忽略 --root, --uid, --gid 的设置" env:"XL_ROOTLESS"`
	Init     bool `usage:"手动处理依赖权限，需要 root 权限执行"`
}

// 配置检查
func ConfigCheck(cfg *Config) (err error) {

	//解决冲突设置
	cfg.Root = utils.Iif(cfg.Rootless || cfg.Init, "/", cfg.Root)

	//设置默认值
	cfg.DashboardPort = cmp.Or(cfg.DashboardPort, 2345) //端口默认2345
	cfg.SpkUrl = cmp.Or(cfg.SpkUrl, SPK_URL)            //下载链接

	cfg.Root = cmp.Or(cfg.Root, "xunlei")                              //根目录默认 ./xunlei
	cfg.DirData = cmp.Or(cfg.DirData, filepath.Join(cfg.Root, "data")) //数据默认 ${root}/data
	if len(cfg.DirDownload) == 0 {
		cfg.DirDownload = []string{filepath.Join(cfg.Root, "downloads")} //下载默认 ${root}/downloads
	}

	//处理路径成绝对路径
	abs, e := fAbs()
	if e != nil {
		return e
	}
	cfg.Root = abs(cfg.Root)
	cfg.DirData = abs(cfg.DirData)
	cfg.DirDownload = pList(cfg.DirDownload, abs)

	cfg.Log = utils.Iif(cfg.Debug || cfg.Init, "debug", cfg.Log)
	return
}

func ConfigPrint(cfg *Config, fPrint func(string)) (cline []string) {
	fPrintf := func(format string, args ...any) {
		fPrint(fmt.Sprintf(format, args...))
	}
	// fPrintf("version: %s", cfg.FlagVersion)
	// fPrintf("buildTime: %s", cfg.FlagBuildTime.In(time.Local).Format(time.RFC3339))

	fPrintf("DASHBOARD_IP: %s", cfg.DashboardIp)
	fPrintf("DASHBOARD_PORT: %d", cfg.DashboardPort)
	fPrintf("DASHBOARD_USERNAME: %s", cfg.DashboardUsername)
	fPrintf("DASHBOARD_PASSWORD: %s", utils.PasswordMask(cfg.DashboardPassword))
	for i, dir := range cfg.DirDownload {
		fPrintf("DIR_DOWNLOAD[%d]: %s", i, dir)
	}
	fPrintf("DIR_DATA: %s", cfg.DirData)
	fPrintf("UID: %d", cfg.Uid)
	fPrintf("UID: %d", cfg.Gid)
	fPrintf("ROOT: %s", cfg.Root)
	fPrintf("SPK: %s", cfg.SpkUrl)
	fPrintf("SPK_FORCE_DOWNLOAD: %t", cfg.SpkForceDownload)
	fPrintf("PREVENT_UPDATE: %t", cfg.PreventUpdate)
	fPrintf("LOG: %s", cfg.Log)
	fPrintf("DEBUG: %t", cfg.Debug)
	return
}

// 生成一个无error返回的绝对路径解析器, (pwd 必须是绝对路径)
func fAbs() (func(string) string, error) {
	wd, e := os.Getwd() //尝试获取当前目录
	if e != nil {
		return nil, fmt.Errorf("get wd fail: %w", e)
	}
	return func(p string) string {
		if filepath.IsAbs(p) {
			return filepath.Clean(p)
		}
		return filepath.Join(wd, p)
	}, nil
}

// 冒号分割路径, 返回分割后的路径的绝对路径数组
func pList(src []string, absFn func(string) string) (dst []string) {
	for _, p := range src {
		for d := range strings.SplitSeq(p, ":") {
			if d = strings.TrimSpace(d); d != "" {
				dst = append(dst, absFn(d))
			}
		}
	}
	return
}
