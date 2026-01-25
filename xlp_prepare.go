package xunlei

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/cnk3x/xunlei/embed/authenticate_cgi"
	"github.com/cnk3x/xunlei/pkg/fo"
	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/cnk3x/xunlei/pkg/vms/sys"
)

// synoPrepare, *root required
//   - 在不存在的情况下创建一系列所需要的文件夹
//   - 在不存在的情况下创建 /etc/synoinfo.conf
//   - 在不存在的情况下创建 /usr/syno/synoman/webman/modules/authenticate.cgi
func synoPrepare(ctx context.Context, cfg *Config) (undo func(), err error) {
	dirs := sys.AddRoot(cfg.Root,
		DIR_SYNOPKG_PKGDEST+"/bin",
		DIR_SYNOPKG_PKGDEST+"/bin/bin",
		DIR_SYNOPKG_PKGDEST+"/ui",
		DIR_SYNOPKG_PKGDEST+"/var",
	)
	dirs = append(dirs, cfg.DirDownload...)
	dirs = append(dirs, cfg.DirData, filepath.Join(cfg.DirData, ".drive"))

	lines := fo.Lines(
		fmt.Sprintf(`platform_name=%q`, SYNO_PLATFORM),
		fmt.Sprintf(`synobios=%q`, SYNO_PLATFORM),
		fmt.Sprintf(`unique=synology_%s_%s`, SYNO_PLATFORM, SYNO_MODEL),
	)

	printTask := func(ctx context.Context, msg string, args ...any) func() (sys.Undo, error) {
		return func() (sys.Undo, error) {
			slog.InfoContext(ctx, msg, args...)
			return nil, nil
		}
	}

	return utils.QRun(
		printTask(ctx, "create dirs ..."),
		sys.MkdirsTask(ctx, append(dirs, DIR_SYNO_MODULES), 0o777),
		printTask(ctx, "chmod ..."),
		sys.ChmodTask(ctx, dirs, 0777, false),
		printTask(ctx, "create files ..."),
		sys.FileCreateTask(ctx, filepath.Join(cfg.Root, FILE_SYNO_AUTHENTICATE_CGI), authenticate_cgi.WriteTo, fo.Perm(0o777), fo.DirPerm(0o777)),
		sys.FileCreateTask(ctx, filepath.Join(cfg.Root, FILE_SYNO_INFO_CONF), lines),
	)
}

// rootless
func synoCheck(cfg *Config) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	path := filepath.Join(cfg.Root, FILE_SYNO_INFO_CONF)
	var validFlag int
	err = fo.OpenRead(path, fo.LineRead(func(s string) error {
		switch {
		case strings.HasPrefix(s, "platform_name="):
			validFlag |= 1
		case strings.HasPrefix(s, "synobios="):
			validFlag |= 2
		case strings.HasPrefix(s, "unique="):
			validFlag |= 4
		}
		return nil
	}))
	if err == nil && validFlag != 1|2|4 {
		err = fmt.Errorf("syno check etc %s: miss some key(platform_name, synobios, unique)", path)
	}
	if err != nil {
		return
	}

	cgi := filepath.Join(cfg.Root, FILE_SYNO_AUTHENTICATE_CGI)
	if err = exec.CommandContext(ctx, cgi).Run(); err != nil {
		if os.IsNotExist(err) {
			err = os.ErrNotExist
		}
		err = fmt.Errorf("syno check usr %s: %w", cgi, err)
	}
	return
}

var _ = synoCheck

// permissionPrint 显示需要的权限已做对比
func permissionPrint(ctx context.Context, cfg *Config) {
	files := sys.AddRoot(cfg.Root,
		DIR_SYNOPKG_PKGDEST+"/bin",
		DIR_SYNOPKG_PKGDEST+"/bin/bin",
		DIR_SYNOPKG_PKGDEST+"/ui",
		DIR_SYNOPKG_PKGDEST+"/var",
		DIR_SYNO_MODULES,
		FILE_SYNO_AUTHENTICATE_CGI,
		FILE_SYNO_INFO_CONF,
	)
	files = append(files, cfg.DirData, filepath.Join(cfg.DirData, ".drive"))
	files = append(files, cfg.DirDownload...)

	slog.DebugContext(ctx, "check permission ...")
	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			slog.DebugContext(ctx, "check permission", "path", file, "err", sys.Errno(err))
		} else {
			stat_t := stat.Sys().(*syscall.Stat_t)
			slog.DebugContext(ctx, "check permission", "perm", sys.Perm2s(stat.Mode()), "uid", stat_t.Uid, "gid", stat_t.Gid, "path", file)
		}
	}
}
