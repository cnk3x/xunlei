package xunlei

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/log"
)

func cmdRun(ctx context.Context, name string, args []string, dir string, env []string, uid, gid uint32) (err error) {
	c := exec.CommandContext(ctx, name, args...)
	c.Dir, c.Env = dir, env
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.Cancel = func() error { return syscall.Kill(-c.Process.Pid, syscall.SIGINT) }
	if uid != 0 || gid != 0 {
		c.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid}
	}

	console := wrapConsole(ctx)
	defer console.Close()
	c.Stdout, c.Stderr = console, console

	if err = c.Start(); err != nil {
		slog.ErrorContext(ctx, "app start fail", "cmd", c.String(), "err", err)
		return
	}
	slog.InfoContext(ctx, "started", "pid", c.Process.Pid, "cmd", c.String())

	err = c.Wait()
	return
}

func wrapConsole(ctx context.Context) io.WriteCloser {
	lv := slog.LevelDebug
	re0 := regexp.MustCompile(`^\d{2}/\d{2} \d{2}:\d{2}:\d{2}(\.\d+)? (INFO|ERROR|WARNING)\s*>?\s*`)
	re1 := regexp.MustCompile(`^[\dTZ:\.-]?\s*(INFO|ERROR|WARNING)\s*(\[\d+\])?\s*`)
	r, w := io.Pipe()
	go func() {
		for scan := bufio.NewScanner(r); scan.Scan(); {
			s := scan.Text()

			if strings.Contains(s, `filter not match`) {
				continue
			}

			if strings.Contains(s, `DetectPlatform err:`) {
				continue
			}

			if strings.Contains(s, `detect err:key file lost`) {
				continue
			}

			switch {
			case strings.HasPrefix(s, "panic:"):
				lv = slog.LevelError
			case strings.HasPrefix(s, "RunSafe panic:"):
				lv = slog.LevelError
			case re0.MatchString(s):
				m := re0.FindStringSubmatch(s)
				if lv = log.LevelFromString(m[2], slog.LevelDebug); lv == slog.LevelInfo {
					lv = slog.LevelDebug
				}
				s = s[len(m[0]):]
			case re1.MatchString(s):
				m := re0.FindStringSubmatch(s)
				if lv = log.LevelFromString(m[1], slog.LevelDebug); lv == slog.LevelInfo {
					lv = slog.LevelDebug
				}
				s = s[len(m[0]):]
			}
			s = strings.ReplaceAll(s, `\u0000`, "")
			slog.Log(ctx, lv, s)
		}
	}()

	return w
}
