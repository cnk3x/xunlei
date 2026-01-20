package xunlei

import (
	"cmp"
	"log/slog"
	"net"
	"path/filepath"

	"github.com/cnk3x/xunlei/pkg/flags"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/cnk3x/xunlei/spk"
)

// Config 配置
type Config struct {
	Port uint16 //网页访问的端口
	Ip   net.IP //网页访问绑定IP，默认绑定所有IP

	DashboardUsername string //网页访问的用户名
	DashboardPassword string //网页访问的密码

	Root             string   //主目录
	DirDownload      []string //下载保存文件夹，可多次指定，需确保有权限访问
	DirData          string   //程序数据保存文件夹，其下'.drive'文件夹中，存储了登录的账号，下载进度等信息
	Uid              int      //运行迅雷的用户ID
	Gid              int      //运行迅雷的用户组ID
	PreventUpdate    bool     //阻止更新
	SpkUrl           string   //下载链接
	SpkForceDownload bool     //是否强制下载

	LauncherLogFile string //迅雷启动器日志文件

	Debug bool   //是否开启调试日志
	Log   string //日志等级 [debug,info/information,warn/warning,error/err]

	chroot string //过期暂留
}

// 默认配置端口2345，下载保存文件夹 /xunlei/downloads, 数据文件夹 /xunlei/data
func ConfigBind(cfg *Config) (err error) {
	cfg.Log = "info"

	flags.Var(&cfg.Port, "dashboard_port", "", "网页访问的端口", "XL_DASHBOARD_PORT", "XL_PORT")
	flags.Var(&cfg.Ip, "dashboard_ip", "", "网页访问绑定IP，默认绑定所有IP", "XL_DASHBOARD_IP", "XL_IP")

	flags.Var(&cfg.DashboardUsername, "dashboard_username", "", "网页访问的用户名", "XL_DASHBOARD_USERNAME", "XL_BA_USER")
	flags.Var(&cfg.DashboardPassword, "dashboard_password", "", "网页访问的密码", "XL_DASHBOARD_PASSWORD", "XL_BA_PASSWORD")

	flags.Var(&cfg.DirDownload, "dir_download", "d", "下载保存文件夹，可多次指定，需确保有权限访问", "XL_DIR_DOWNLOAD")
	flags.Var(&cfg.DirData, "dir_data", "D", "程序数据保存文件夹，其下'.drive'文件夹中，存储了登录的账号，下载进度等信息", "XL_DIR_DATA")
	flags.Var(&cfg.Uid, "uid", "u", "运行迅雷的用户ID", "XL_UID", "UID")
	flags.Var(&cfg.Gid, "gid", "g", "运行迅雷的用户组ID", "XL_GID", "GID")
	flags.Var(&cfg.PreventUpdate, "prevent_update", "", "阻止更新", "XL_PREVENT_UPDATE")
	flags.Var(&cfg.Root, "root", "r", "主目录", "XL_ROOT", "XL_CHROOT")
	flags.Var(&cfg.SpkUrl, "spk", "", "SPK 下载链接", "XL_SPK")
	flags.Var(&cfg.SpkForceDownload, "spk_force_download", "F", "强制下载", "XL_SPK_FORCE_DOWNLOAD")
	flags.Var(&cfg.LauncherLogFile, "launcher_log_file", "", "迅雷启动器日志文件", "XL_LAUNCHER_LOG_FILE")
	flags.Var(&cfg.Debug, "debug", "", "调试模式,调试模式下，失败也不会自动退出,在此模式行，日志等级自动调整为`debug`", "XL_DEBUG")
	flags.Var(&cfg.Log, "log", "", "日志等级, 可选 debug,info/information,warn/warning,error/err", "XL_LOG")
	flags.Var(&cfg.chroot, "chroot", "", "已过期 **请使用 --root/-r 替代**")

	if err = flags.Parse(); err != nil {
		return
	}

	log.ForDefault(utils.Iif(cfg.Debug, "debug", cfg.Log), false)

	cfg.Port = cmp.Or(cfg.Port, 2345)
	cfg.SpkUrl = cmp.Or(cfg.SpkUrl, spk.DownloadUrl)

	if cfg.Root, err = filepath.Abs(cmp.Or(cfg.Root, cfg.chroot, "xunlei")); err != nil {
		slog.Error("无法获取绝对路径", "root", cfg.Root, "err", err)
		return
	}

	cfg.DirData = cmp.Or(cfg.DirData, filepath.Join(cfg.Root, "data"))
	if cfg.DirData, err = filepath.Abs(cfg.DirData); err != nil {
		return
	}

	if len(cfg.DirDownload) == 0 {
		cfg.DirDownload = utils.Array(filepath.Join(cfg.Root, "downloads"))
	}

	if cfg.DirDownload, err = utils.Replace(cfg.DirDownload, filepath.Abs); err != nil {
		slog.Error("无法获取绝对路径", "dir_download", cfg.DirDownload, "err", err)
		return
	}
	return
}
