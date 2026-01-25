package xunlei

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/vms"
	"github.com/cnk3x/xunlei/spk"
)

const Version = "4.0.0-beta1"

const (
	SYNOPKG_DSM_VERSION_MAJOR = "7"                                                                                     //系统的主版本
	SYNOPKG_DSM_VERSION_MINOR = "2"                                                                                     //系统的次版本
	SYNOPKG_DSM_VERSION_BUILD = "64570"                                                                                 //系统的编译版本
	SYNOPKG_PKGNAME           = "pan-xunlei-com"                                                                        //包名
	DIR_SYNOPKG_PKGROOT       = "/var/packages/pan-xunlei-com"                                                          //包安装目录
	DIR_SYNOPKG_PKGDEST       = "/var/packages/pan-xunlei-com/target"                                                   //包安装目录
	DIR_SYNOPKG_WORK          = "/var/packages/pan-xunlei-com/bin"                                                      //
	FILE_PAN_XUNLEI_VER       = "/var/packages/pan-xunlei-com/target/bin/bin/version"                                   //版本文件
	FILE_PAN_XUNLEI_CLI       = "/var/packages/pan-xunlei-com/target/bin/bin/xunlei-pan-cli-launcher." + runtime.GOARCH //启动器
	FILE_INDEX_CGI            = "/var/packages/pan-xunlei-com/target/ui/index.cgi"                                      //CGI文件路径
	DIR_VAR                   = "/var/packages/pan-xunlei-com/target/var"                                               //SYNOPKG_PKGROOT
	FILE_PID                  = "/var/packages/pan-xunlei-com/target/var/pan-xunlei-com.pid"                            //进程文件
	SOCK_LAUNCHER_LISTEN      = "/var/packages/pan-xunlei-com/target/var/pan-xunlei-com-launcher.sock"                  //启动器监听地址
	SOCK_DRIVE_LISTEN         = "/var/packages/pan-xunlei-com/target/var/pan-xunlei-com.sock"                           //主程序监听地址
	FILE_SYNO_INFO_CONF       = "/etc/synoinfo.conf"                                                                    //synoinfo.conf 文件路径

	DIR_SYNO_MODULES           = "/usr/syno/synoman/webman/modules/"
	FILE_SYNO_AUTHENTICATE_CGI = "/usr/syno/synoman/webman/modules/authenticate.cgi" //syno...authenticate.cgi 文件路径

	SYNO_VERSION = SYNO_PLATFORM + " dsm " + SYNOPKG_DSM_VERSION_MAJOR + "." + SYNOPKG_DSM_VERSION_MINOR + "-" + SYNOPKG_DSM_VERSION_BUILD //系统版本
)

func Run(ctx context.Context, cfg *Config) error {
	isRootUser := os.Getuid() == 0              //当前用户是否root
	rootRequired := cfg.Root != "/" || cfg.Init //当前方式是否需要root

	if !isRootUser {
		if cfg.Init {
			return fmt.Errorf("root required, cause: init=true")
		}
		if cfg.Root != "/" {
			return fmt.Errorf("root required, cause: root=%s", cfg.Root)
		}
	}

	if rootRequired {
		undo, err := synoPrepare(log.Prefix(ctx, "init"), cfg)
		if err != nil {
			return err
		}
		if !cfg.Init {
			defer undo()
		}
	}

	permissionPrint(log.Prefix(ctx, "prep"), cfg)

	if cfg.Init {
		return nil
	}

	if cfg.Rootless {
		return launch(cfg)(log.Prefix(ctx, "lrun"))
	}

	return vms.Execute(
		log.Prefix(ctx, "rrun"),
		vms.Wait(cfg.Debug),
		vms.Root(cfg.Root),
		vms.User(cfg.Uid, cfg.Gid),
		vms.Binds(cfg.DirData),
		vms.Binds(cfg.DirDownload...),
		vms.Binds("/lib", "/bin", "/etc/ssl", "/usr"),
		vms.Links("/etc/timezone", "/etc/localtime", "/etc/resolv.conf"),
		vms.Links("/etc/passwd", "/etc/group", "/etc/shadow"),
		vms.Symlink("lib", filepath.Join(cfg.Root, "lib64")),
		vms.Basic,
		vms.Run(launch(cfg)),
	)
}

func launch(cfg *Config) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		if err := spk.Download(ctx, cfg.SpkUrl, DIR_SYNOPKG_PKGDEST, false); err != nil {
			return err
		}

		cDone, err := coreRun(log.Prefix(ctx, "core"), cfg)
		if err != nil {
			return err
		}

		wDone, err := dashboardRun(log.Prefix(ctx, "http"), cfg)
		if err != nil {
			return err
		}

		go func() {
			select {
			case <-ctx.Done():
				return
			case <-cDone:
			case <-wDone:
			}
			cancel()
		}()

		<-wDone
		<-cDone
		return nil
	}
}

func mockEnv(dirData, dirDownload string) []string {
	// ld_lib := os.Getenv("LD_LIBRARY_PATH")
	return append(os.Environ(),
		"SYNOPLATFORM="+SYNO_PLATFORM,
		"SYNOPKG_PKGNAME="+SYNOPKG_PKGNAME,
		"SYNOPKG_PKGDEST="+DIR_SYNOPKG_PKGDEST,
		"SYNOPKG_DSM_VERSION_MAJOR="+SYNOPKG_DSM_VERSION_MAJOR,
		"SYNOPKG_DSM_VERSION_MINOR="+SYNOPKG_DSM_VERSION_MINOR,
		"SYNOPKG_DSM_VERSION_BUILD="+SYNOPKG_DSM_VERSION_BUILD,
		"DriveListen=unix://"+SOCK_DRIVE_LISTEN,
		"PLATFORM=群晖",
		"OS_VERSION="+SYNO_VERSION,
		"ConfigPath="+dirData,
		"HOME="+filepath.Join(dirData, ".drive"),
		"DownloadPATH="+dirDownload,
		"GIN_MODE=release",
		// "LD_LIBRARY_PATH=/lib"+utils.Iif(ld_lib == "", "", ":")+ld_lib,
	)
}

// NewnsRun 在unshare后执行 fn，执行完成后恢复
//
// 参数:
//   - ctx: 上下文
//   - fn: 要在新命名空间中执行的函数
func NewnsRun(ctx context.Context, fn func()) {
	// 创建新的挂载命名空间、PID命名空间和UTS命名空间
	syscall.Unshare(syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS)
	// 设置根目录为私有挂载，防止影响父命名空间
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	// 挂载新的proc文件系统
	syscall.Mount("none", "/proc", "proc", syscall.MS_NOSUID|syscall.MS_NOEXEC|syscall.MS_NODEV, "")
	// 执行指定函数
	fn()
	// 清理操作：卸载挂载的proc文件系统
	syscall.Unmount("/proc", syscall.MNT_DETACH)
	// 注意：由于CLONE_NEWPID的存在，子进程退出后会自动清理PID命名空间
	// 挂载命名空间会在进程结束时自动恢复到原始状态
	// 但由于我们设置了MS_PRIVATE，所以不会影响原始挂载树
}
