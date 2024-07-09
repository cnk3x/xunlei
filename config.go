package xunlei

import (
	"net"
	"path/filepath"

	"github.com/cnk3x/xunlei/pkg/flags"
	"github.com/cnk3x/xunlei/pkg/lod"
)

// Config 配置
type Config struct {
	Port              uint16         `flag:"dashboard_port,p"       env:"XL_DASHBOARD_PORT"                    usage:"网页访问的端口"`
	Ip                net.IP         `flag:"dashboard_ip,i"         env:"XL_DASHBOARD_IP"                      usage:"网页访问绑定IP，默认绑定所有IP"`
	DashboardUsername string         `flag:"dashboard_username,u"   env:"XL_DASHBOARD_USERNAME,XL_BA_USER"     usage:"网页访问的用户名"`
	DashboardPassword string         `flag:"dashboard_password,k"   env:"XL_DASHBOARD_PASSWORD,XL_BA_PASSWORD" usage:"网页访问的密码"`
	DirDownload       flags.PathList `flag:"dir_download"           env:"XL_DIR_DOWNLOAD"                      usage:"下载保存文件夹，可多次指定，需确保有权限访问"`
	DirData           string         `flag:"dir_data"               env:"XL_DIR_DATA"                          usage:"程序数据保存文件夹，其下'.drive'文件夹中，存储了登录的账号，下载进度等信息"`
	Uid               uint32         `flag:"uid"                    env:"XL_UID,UID"                           usage:"运行迅雷的用户ID"`
	Gid               uint32         `flag:"gid"                    env:"XL_GID,GID"                           usage:"运行迅雷的用户组ID"`
	PreventUpdate     bool           `flag:"prevent_update"         env:"XL_PREVENT_UPDATE" default:"true"     usage:"阻止更新"`
	Chroot            string         `flag:"chroot,r"               env:"XL_CHROOT"                            usage:"CHROOT主目录, 指定该值则以chroot模式运行, 这用于在容器内隔离环境"`
	Debug             bool           `flag:"debug"                  env:"XL_DEBUG"                             usage:"是否开启调试日志"`
}

func ConfigDefault() (cfg Config) {
	cfg.SetDefault()
	return
}

// SetDefault 默认配置端口2345，下载保存文件夹 /xunlei/downloads, 数据文件夹 /xunlei/data
func (cfg *Config) SetDefault() {
	cfg.Port = lod.Select(cfg.Port, 2345)
	cfg.Chroot = lod.Select(cfg.Chroot, "/")
	cfg.DirData = lod.Select(cfg.DirData, "/xunlei/data")
	cfg.DirDownload = cfg.DirDownload.IfNil("/xunlei/downloads")
}

func (cfg *Config) Validate() (err error) {
	cfg.SetDefault()

	cfg.DirDownload = cfg.DirDownload.Abs()

	if cfg.DirData != "" {
		if cfg.DirData, err = filepath.Abs(cfg.DirData); err != nil {
			return
		}
	}

	if cfg.Chroot != "" {
		if cfg.Chroot, err = filepath.Abs(cfg.Chroot); err != nil {
			return
		}
	}

	return
}
