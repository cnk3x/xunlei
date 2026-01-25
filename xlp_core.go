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

	"github.com/cnk3x/xunlei/pkg/cmdx"
	"github.com/cnk3x/xunlei/pkg/fo"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/vms/sys"
)

func coreRun(ctx context.Context, cfg *Config) (done <-chan struct{}, err error) {
	envs := mockEnv(cfg.DirData, strings.Join(cfg.DirDownload, ":"))
	panCtx := log.Prefix(ctx, "pan")

	c := exec.CommandContext(ctx, FILE_PAN_XUNLEI_CLI,
		"-launcher_listen", "unix://"+SOCK_LAUNCHER_LISTEN, "-pid", FILE_PID,
	)
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.Cancel = func() error { return syscall.Kill(-c.Process.Pid, syscall.SIGINT) }

	if cfg.PreventUpdate {
		c.Args = append(c.Args, "-update_url", "null")
	}

	if cfg.LauncherLogFile != "" {
		c.Args = append(c.Args, "-logfile", cfg.LauncherLogFile)
	}

	c.Dir = DIR_SYNOPKG_WORK
	c.Env = envs

	sw := cmdx.LineWriter(func(s string) { slog.DebugContext(panCtx, "[std] "+s) })
	ew := cmdx.LineWriter(func(s string) { slog.DebugContext(panCtx, "[err] "+s) })
	c.Stdout, c.Stderr = sw, ew
	defer sw.Close()
	defer ew.Close()

	cStarted := make(chan error, 1)
	cDone := make(chan struct{})
	done = cDone

	go func() {
		defer close(cStarted)
		defer close(cDone)

		cleanExit(DIR_VAR)
		defer cleanExit(DIR_VAR)
		if err = c.Start(); err != nil {
			slog.DebugContext(ctx, "core start", "cmdline", c.String(), "dir", c.Dir, "err", sys.Errno(err))
			cStarted <- err
			return
		}
		slog.DebugContext(ctx, "core start", "cmdline", c.String(), "dir", c.Dir, "pid", c.Process.Pid)
		cStarted <- nil

		if err = c.Wait(); err != nil {
			if err == context.Canceled {
				err = nil
			} else if signal := cmdx.GetSignal(err); signal != -1 {
				slog.DebugContext(ctx, "core done", "signal", fmt.Sprintf("%s(%d)", signal, signal))
			} else {
				slog.DebugContext(ctx, "core done", "err", sys.Errno(err))
			}
			return
		}
		slog.DebugContext(ctx, "core done")
	}()

	err = <-cStarted
	return
}

func cleanExit(dir string) {
	for name := range fo.FileSeq(dir, fo.MatchExt(".sock", ".pid")) {
		os.Remove(filepath.Join(dir, name))
	}
}
