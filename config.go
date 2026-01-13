package xunlei

import (
	"cmp"
	"net"
	"path/filepath"

	"github.com/cnk3x/xunlei/pkg/flags"
	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/cnk3x/xunlei/spk"
)

// Config 配置
type Config struct {
	Debug bool //是否开启调试日志

	Port uint16 //网页访问的端口
	Ip   net.IP //网页访问绑定IP，默认绑定所有IP

	DashboardUsername string //网页访问的用户名
	DashboardPassword string //网页访问的密码

	DirDownload   []string //下载保存文件夹，可多次指定，需确保有权限访问
	DirData       string   //程序数据保存文件夹，其下'.drive'文件夹中，存储了登录的账号，下载进度等信息
	Uid           uint32   //运行迅雷的用户ID
	Gid           uint32   //运行迅雷的用户组ID
	PreventUpdate bool     //阻止更新
	Chroot        string   //CHROOT主目录, 指定该值则以chroot模式运行, 这用于在容器内隔离环境
	SpkUrl        string   //
	ForceDownload bool     //是否强制下载

	LauncherLogFile string
}

// 默认配置端口2345，下载保存文件夹 /xunlei/downloads, 数据文件夹 /xunlei/data
func ConfigBind(cfg *Config) (err error) {
	cfg.Port = 2345
	cfg.Chroot = "./xunlei"
	cfg.DirData = "./xunlei/data"
	cfg.DirDownload = []string{"./xunlei/downloads"}
	cfg.SpkUrl = spk.DownloadUrl

	flags.Var(&cfg.Port, "dashboard_port", "", "网页访问的端口", "XL_DASHBOARD_PORT", "XL_PORT")
	flags.Var(&cfg.Ip, "dashboard_ip", "", "网页访问绑定IP，默认绑定所有IP", "XL_DASHBOARD_IP", "XL_IP")

	flags.Var(&cfg.DashboardUsername, "dashboard_username", "", "网页访问的用户名", "XL_DASHBOARD_USERNAME", "XL_BA_USER")
	flags.Var(&cfg.DashboardPassword, "dashboard_password", "", "网页访问的密码", "XL_DASHBOARD_PASSWORD", "XL_BA_PASSWORD")

	flags.Var(&cfg.DirDownload, "dir_download", "d", "下载保存文件夹，可多次指定，需确保有权限访问", "XL_DIR_DOWNLOAD")
	flags.Var(&cfg.DirData, "dir_data", "D", "程序数据保存文件夹，其下'.drive'文件夹中，存储了登录的账号，下载进度等信息", "XL_DIR_DATA")
	flags.Var(&cfg.Uid, "uid", "u", "运行迅雷的用户ID", "XL_UID", "UID")
	flags.Var(&cfg.Gid, "gid", "g", "运行迅雷的用户组ID", "XL_GID", "GID")
	flags.Var(&cfg.PreventUpdate, "prevent_update", "", "阻止更新", "XL_PREVENT_UPDATE")
	flags.Var(&cfg.Chroot, "chroot", "r", "CHROOT主目录", "XL_CHROOT")
	flags.Var(&cfg.SpkUrl, "spk", "", "SPK 下载链接", "XL_SPK_URL")
	flags.Var(&cfg.ForceDownload, "force_download", "F", "强制下载")
	flags.Var(&cfg.Debug, "debug", "", "是否开启调试日志", "XL_DEBUG")

	flags.Var(&cfg.LauncherLogFile, "launcher_log_file", "", "迅雷启动器日志")

	if err = flags.Parse(); err != nil {
		return
	}

	for i, dir := range cfg.DirDownload {
		if cfg.DirDownload[i], err = filepath.Abs(dir); err != nil {
			return
		}
	}

	if cfg.DirDownload = utils.CompactUniq(cfg.DirDownload); len(cfg.DirDownload) == 0 {
		cfg.DirDownload = []string{"/xunlei/downloads"}
	}

	if cfg.DirData, err = filepath.Abs(cmp.Or(cfg.DirData, "/xunlei/data")); err != nil {
		return
	}

	if cfg.Chroot, err = filepath.Abs(cmp.Or(cfg.Chroot, "/")); err != nil {
		return
	}

	cfg.SpkUrl = cmp.Or(cfg.SpkUrl, spk.DownloadUrl)
	return
}
