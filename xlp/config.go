package xlp

import (
	"fmt"
	"os"
)

const (
	ENV_DASHBOARD_PORT     = "XL_DASHBOARD_PORT"     //环境变量名称: 网页访问的端口
	ENV_DASHBOARD_USERNAME = "XL_DASHBOARD_USERNAME" //环境变量名称: 网页访问的用户名
	ENV_DASHBOARD_PASSWORD = "XL_DASHBOARD_PASSWORD" //环境变量名称: 网页访问的密码
	ENV_DIR_DOWNLOAD       = "XL_DIR_DOWNLOAD"       //环境变量名称: 下载保存文件夹
	ENV_DIR_DATA           = "XL_DIR_DATA"           //环境变量名称: 程序数据保存文件夹
	ENV_XL_UID             = "XL_UID"                //运行迅雷的用户ID
	ENV_XL_GID             = "XL_GID"                //运行迅雷的用户组ID
	ENV_XL_DEBUG           = "XL_DEBUG"              //是否开启调试日志
	ENV_CHROOT             = "XL_CHROOT"             //CHROOT模式运行，用于在容器内。
	ENV_PREVENT_UPDATE     = "XL_PREVENT_UPDATE"     //阻止更新

	ENV_WEB_ADDRESS = "XL_WEB_ADDRESS" //旧的环境变量，标记过期: 网页访问的地址
	ENV_BA_USER     = "XL_BA_USER"     //旧的环境变量，标记过期: 网页访问的用户名
	ENV_BA_PASSWORD = "XL_BA_PASSWORD" //旧的环境变量，标记过期: 网页访问的密码

	ENV_OLD_UID = "UID"
	ENV_OLD_GID = "GID"
)

// 配置
type Config struct {
	DashboardPort     int    `json:"dashboard_port,omitempty"`     //网页访问的端口
	DashboardUsername string `json:"dashboard_username,omitempty"` //网页访问的用户名
	DashboardPassword string `json:"dashboard_password,omitempty"` //网页访问的密码
	DirDownload       string `json:"dir_download,omitempty"`       //下载保存文件夹，此文件夹链接到主文件夹的`下载`
	DirData           string `json:"dir_data,omitempty"`           //程序数据保存文件夹，此文件夹链接到主文件夹的`.drive，存储了登录的账号，下载进度等信息
	Debug             bool   `json:"debug,omitempty"`              //是否开启调试日志
	UID               string `json:"uid,omitempty"`                //运行迅雷的用户ID
	GID               string `json:"gid,omitempty"`                //运行迅雷的用户组ID
	PreventUpdate     bool   `json:"prevent_update,omitempty"`     //阻止更新
	Chroot            string `json:"chroot,omitempty"`             //CHROOT模式运行，用于在容器内。
}

// 绑定环境变量，参数来源和优先级： 命令行参数 > 环境变量 > 预设
func (cfg *Config) Init() {
	bindEnv(&cfg.DashboardPort, ENV_DASHBOARD_PORT, ENV_DASHBOARD_PORT)
	bindEnv(&cfg.DashboardUsername, ENV_DASHBOARD_USERNAME, ENV_BA_USER)
	bindEnv(&cfg.DashboardPassword, ENV_DASHBOARD_PASSWORD, ENV_BA_PASSWORD)
	bindEnv(&cfg.DirDownload, ENV_DIR_DOWNLOAD)
	bindEnv(&cfg.DirData, ENV_DIR_DATA)
	bindEnv(&cfg.Debug, ENV_XL_DEBUG)
	bindEnv(&cfg.UID, ENV_XL_UID, ENV_OLD_UID)
	bindEnv(&cfg.GID, ENV_XL_GID, ENV_OLD_GID)
	bindEnv(&cfg.Chroot, ENV_CHROOT)
	bindEnv(&cfg.PreventUpdate, ENV_PREVENT_UPDATE)
}

func (cfg *Config) MarshalEnv() (env []string) {
	env = append(env, fmt.Sprintf("%s=%d", ENV_DASHBOARD_PORT, cfg.DashboardPort))
	env = append(env, fmt.Sprintf("%s=%s", ENV_DASHBOARD_USERNAME, cfg.DashboardUsername))
	env = append(env, fmt.Sprintf("%s=%s", ENV_DASHBOARD_PASSWORD, cfg.DashboardPassword))
	env = append(env, fmt.Sprintf("%s=%s", ENV_DIR_DOWNLOAD, cfg.DirDownload))
	env = append(env, fmt.Sprintf("%s=%s", ENV_DIR_DATA, cfg.DirData))
	env = append(env, fmt.Sprintf("%s=%t", ENV_XL_DEBUG, cfg.Debug))
	env = append(env, fmt.Sprintf("%s=%s", ENV_XL_UID, cfg.UID))
	env = append(env, fmt.Sprintf("%s=%s", ENV_XL_GID, cfg.GID))
	env = append(env, fmt.Sprintf("%s=%t", ENV_PREVENT_UPDATE, cfg.PreventUpdate))
	return
}

// 绑定命令行参数，参数来源和优先级： 命令行参数 > 环境变量 > 预设
func (cfg *Config) BindFlag(fs flagSet, parse bool) {
	fs.IntVar(&cfg.DashboardPort, "dashboard_port", cfg.DashboardPort, "网页控制台访问端口")
	fs.StringVar(&cfg.DashboardUsername, "dashboard_username", cfg.DashboardUsername, "网页控制台访问用户名")
	fs.StringVar(&cfg.DashboardPassword, "dashboard_password", cfg.DashboardPassword, "网页控制台访问密码")
	fs.StringVar(&cfg.DirDownload, "dir_download", cfg.DirDownload, "默认下载保存文件夹")
	fs.StringVar(&cfg.DirData, "dir_data", cfg.DirData, "迅雷程序数据保存文件夹")
	fs.StringVar(&cfg.UID, "uid", cfg.UID, "运行迅雷的 UID")
	fs.StringVar(&cfg.GID, "gid", cfg.GID, "运行迅雷的 GID")
	fs.BoolVar(&cfg.Debug, "debug", cfg.Debug, "开启调试模式")
	fs.StringVar(&cfg.Chroot, "chroot", cfg.Chroot, "CHROOT模式运行，用于在容器内。")
	fs.BoolVar(&cfg.PreventUpdate, "prevent_update", cfg.PreventUpdate, "CHROOT模式运行，用于在容器内。")

	if parse {
		fs.Parse(os.Args[1:])
	}
}

func (cfg Config) UserMap() map[string]string {
	return map[string]string{
		cfg.DashboardUsername: cfg.DashboardPassword,
	}
}

type flagSet interface {
	BoolVar(p *bool, name string, value bool, usage string)
	StringVar(p *string, name string, value string, usage string)
	IntVar(p *int, name string, value int, usage string)
	Parse(arguments []string) error
}
