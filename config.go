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
	cfg.Chroot = utils.Eon(filepath.Abs("./xunlei"))
	cfg.DirData = utils.Eon(filepath.Abs("./xunlei/data"))
	cfg.DirDownload = []string{utils.Eon(filepath.Abs("./xunlei/downloads"))}
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
	flags.Var(&cfg.SpkUrl, "spk", "", "SPK 下载链接", "XL_SPK")
	flags.Var(&cfg.ForceDownload, "force_download", "F", "强制下载", "XL_SPK_FORCE_DOWNLOAD")
	flags.Var(&cfg.LauncherLogFile, "launcher_log_file", "", "迅雷启动器日志", "XL_LAUNCHER_LOG_FILE")
	flags.Var(&cfg.Debug, "debug", "", "是否开启调试日志", "XL_DEBUG")

	if err = utils.SeqExec(
		flags.Parse,
		checkPath(&cfg.Chroot, "./xunlei"),
		checkPath(&cfg.DirData, "./xunlei/data"),
		checkPaths(&cfg.DirDownload, "./xunlei/downloads"),
	); err != nil {
		return
	}
	cfg.SpkUrl = cmp.Or(cfg.SpkUrl, spk.DownloadUrl)
	return
}

// 检查路径，并把传入的路径转换成绝对路径
func checkPaths(dirs *[]string, defPath string) func() (err error) {
	return func() (err error) {
		*dirs = utils.CompactUniq(*dirs)
		if len(*dirs) == 0 {
			*dirs = utils.Array(defPath)
		}
		for i, dir := range *dirs {
			if (*dirs)[i], err = filepath.Abs(dir); err != nil {
				*dirs = (*dirs)[:0]
				return
			}
		}
		*dirs = utils.CompactUniq(*dirs)
		return
	}
}

// 检查路径，并把传入的路径转换成绝对路径
func checkPath(path *string, defPath string) func() (err error) {
	return func() (err error) {
		p, e := filepath.Abs(cmp.Or(*path, defPath))
		if err = e; err != nil {
			return
		}
		*path = p
		return
	}
}
