package xlp

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

const (
	ENV_DASHBOARD_PORT     = "XL_DASHBOARD_PORT"     //环境变量名称: 网页访问的端口
	ENV_DASHBOARD_HOST     = "XL_DASHBOARD_HOST"     //环境变量名称: 网页访问的地址
	ENV_DASHBOARD_USER     = "XL_DASHBOARD_USERNAME" //环境变量名称: 网页访问的用户名
	ENV_DASHBOARD_PASSWORD = "XL_DASHBOARD_PASSWORD" //环境变量名称: 网页访问的密码
	ENV_DIR_DOWNLOAD       = "XL_DIR_DOWNLOAD"       //环境变量名称: 下载保存文件夹
	ENV_DIR_DATA           = "XL_DIR_DATA"           //环境变量名称: 程序数据保存文件夹
	ENV_LOG_TARGET         = "XL_LOG"                //环境变量名称: 是否开启调试日志
	ENV_LOG_MAXSIZE        = "XL_LOG_MAXSIZE"        //环境变量名称: 日志文件最大大小
	ENV_LOG_COMPRESS       = "XL_LOG_COMPRESS"       //环境变量名称: 是否压缩日志文件

	ENV_WEB_ADDRESS = "XL_WEB_ADDRESS" //旧的环境变量，标记过期: 网页访问的地址
	ENV_DEBUG       = "XL_DEBUG"       //旧的环境变量，标记过期: 是否开启调试日志
	ENV_BA_USER     = "XL_BA_USER"     //旧的环境变量，标记过期: 网页访问的用户名
	ENV_BA_PASSWORD = "XL_BA_PASSWORD" //旧的环境变量，标记过期: 网页访问的密码
)

// 配置
type Config struct {
	DashboardPort     int    `json:"dashboard_port,omitempty"`     //网页访问的端口
	DashboardHost     string `json:"dashboard_host,omitempty"`     //网页访问的地址
	DashboardUsername string `json:"dashboard_username,omitempty"` //网页访问的用户名
	DashboardPassword string `json:"dashboard_password,omitempty"` //网页访问的密码
	DirDownload       string `json:"dir_download,omitempty"`       //下载保存文件夹，此文件夹链接到主文件夹的`下载`
	DirData           string `json:"dir_data,omitempty"`           //程序数据保存文件夹，此文件夹链接到主文件夹的`.drive，存储了登录的账号，下载进度等信息
	Logger            Logger `json:"logger,omitempty"`
}

// 日志配置
type Logger struct {
	Target   string `json:"target,omitempty"`   // 调用程序的日志输出目标，默认为 null, 可选 file, console
	Maxsize  string `json:"maxsize,omitempty"`  // rotate file when gatter than this value
	Compress bool   `json:"compress,omitempty"` // compress backup file
	Debug    bool   `json:"debug,omitempty"`    // 调试模式（兼容旧版配置） = target: "stderr"
}

// 绑定命令行参数，参数来源和优先级： 命令行参数 > 环境变量 > 预设
func (cfg *Config) Flag(fs flagSet) {
	fs.IntVar(&cfg.DashboardPort, "dashboard-port", cfg.DashboardPort, "网页控制台访问端口")
	fs.StringVar(&cfg.DashboardHost, "dashboard-host", cfg.DashboardHost, "网页控制台访问绑定主机或IP, 不明白留空即可")
	fs.StringVar(&cfg.DashboardUsername, "dashboard-user", cfg.DashboardUsername, "网页控制台访问用户名")
	fs.StringVar(&cfg.DashboardPassword, "dashboard-password", cfg.DashboardPassword, "网页控制台访问密码")
	fs.StringVar(&cfg.DirDownload, "dir-download", cfg.DirDownload, "默认下载保存文件夹")
	fs.StringVar(&cfg.DirData, "dir-data", cfg.DirData, "迅雷程序数据保存文件夹")
	fs.StringVar(&cfg.Logger.Target, "log", cfg.Logger.Target, "日志输出位置, 可选 null, console, file")
	fs.BoolVar(&cfg.Logger.Compress, "log-compress", cfg.Logger.Compress, "日志文件是否压缩")
	fs.StringVar(&cfg.Logger.Maxsize, "log-maxsize", cfg.Logger.Maxsize, "日志文件最大大小")
}

// 绑定环境变量，参数来源和优先级： 命令行参数 > 环境变量 > 预设
func (cfg *Config) bindEnv() {
	bindEnv := func(out any, keys ...string) {
		s, found := func(keys ...string) (string, bool) {
			for i, key := range keys {
				if v := os.Getenv(key); v != "" {
					if i > 0 {
						fmt.Printf("[WARN] 环境变量参数%q已过期,请使用%q替代\n", key, keys[0])
					} else if key == ENV_DEBUG {
						fmt.Printf("[WARN] 环境变量参数%q已过期,请使用%q替代\n", ENV_DEBUG, ENV_LOG_TARGET)
					}
					return v, true
				}
			}
			return "", false
		}(keys...)

		if !found {
			return
		}

		switch n := out.(type) {
		case *string:
			*n = s
		default:
			switch n := n.(type) {
			case *bool:
				if r, e := strconv.ParseBool(s); e == nil {
					*n = r
				}
			case *int:
				if r, e := strconv.Atoi(s); e == nil {
					*n = r
				}
			case *int64:
				if r, e := strconv.ParseInt(s, 0, 0); e == nil {
					*n = r
				}
			case *uint:
				if r, e := strconv.ParseUint(s, 0, 0); e == nil {
					*n = uint(r)
				}
			case *uint64:
				if r, e := strconv.ParseUint(s, 0, 0); e == nil {
					*n = r
				}
			case *float64:
				if r, e := strconv.ParseFloat(s, 64); e == nil {
					*n = r
				}
			case *float32:
				if r, e := strconv.ParseFloat(s, 32); e == nil {
					*n = float32(r)
				}
			default:
				panic(fmt.Sprintf("unsupported type: %T: %#v", out, out))
			}
		}
	}

	bindEnv(&cfg.DashboardPort, ENV_DASHBOARD_PORT, ENV_DASHBOARD_PORT)
	bindEnv(&cfg.DashboardHost, ENV_DASHBOARD_HOST, ENV_WEB_ADDRESS)
	bindEnv(&cfg.DashboardUsername, ENV_DASHBOARD_USER, ENV_BA_USER)
	bindEnv(&cfg.DashboardPassword, ENV_DASHBOARD_PASSWORD, ENV_BA_PASSWORD)
	bindEnv(&cfg.DirDownload, ENV_DIR_DOWNLOAD)
	bindEnv(&cfg.DirData, ENV_DIR_DATA)
	bindEnv(&cfg.Logger.Target, ENV_LOG_TARGET)
	bindEnv(&cfg.Logger.Compress, ENV_LOG_COMPRESS)
	bindEnv(&cfg.Logger.Maxsize, ENV_LOG_MAXSIZE)
	bindEnv(&cfg.Logger.Debug, ENV_DEBUG)
}

// 填充默认值
func (cfg *Config) fill() {
	cfg.DashboardPort = nSelect(cfg.DashboardPort, 2345)
	cfg.DirDownload = nSelect(cfg.DirDownload, "/downloads")
	cfg.DirData = nSelect(cfg.DirData, "/data")
	cfg.Logger.Target = nSelect(cfg.Logger.Target, Iif(cfg.Logger.Debug, "stderr", "null"))
	cfg.Logger.Maxsize = nSelect(cfg.Logger.Maxsize, "5M")
}

func (cfg Config) UserMap() map[string]string {
	return map[string]string{
		cfg.DashboardUsername: cfg.DashboardPassword,
	}
}

func (l Logger) Create(path string) io.WriteCloser {
	switch l.Target {
	case "file":
		return Rotate(path, l.Maxsize, l.Compress)
	case "console", "std", "stdout":
		return nopcw{os.Stdout}
	case "stderr":
		return nopcw{os.Stderr}
	default:
		return nopcw{io.Discard}
	}
}

type flagSet interface {
	BoolVar(p *bool, name string, value bool, usage string)
	StringVar(p *string, name string, value string, usage string)
	IntVar(p *int, name string, value int, usage string)
}
