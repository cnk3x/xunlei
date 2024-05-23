package xlp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unicode"
)

func loggerRedirect(ctx context.Context, c *exec.Cmd, debug bool) (err error) {
	var epr, spr io.ReadCloser

	c.Stderr = nil
	if epr, err = c.StderrPipe(); err != nil {
		return
	}

	c.Stdout = nil
	if spr, err = c.StdoutPipe(); err != nil {
		return
	}

	name := filepath.Base(c.Path)
	slog.Info(name+" start", "cmd", c.String())

	if err = c.Start(); err == nil {
		slog.Info(name+" started", "pid", c.Process.Pid)

		readLog := func(r io.Reader, app string, debug bool) GroupCallFunc {
			return func(ctx context.Context) (err error) {
				var msg string
				var l = slog.LevelDebug
				var trimFn = func(r rune) bool { return unicode.IsSpace(r) || r == '>' }

				for br := bufio.NewScanner(r); br.Scan(); {
					s := br.Text()
					switch {
					case ctx.Err() != nil:
						return
					case strings.Contains(s, "ERROR"):
						_, msg, _ = strings.Cut(s, "ERROR")
						msg = strings.TrimLeftFunc(msg, trimFn)
						l = slog.LevelError
					case strings.Contains(s, "WARNING"):
						_, msg, _ = strings.Cut(s, "WARNING")
						msg = strings.TrimLeftFunc(msg, trimFn)
						l = slog.LevelWarn
					case debug && strings.Contains(s, "INFO"):
						_, msg, _ = strings.Cut(s, "INFO")
						msg = strings.TrimLeftFunc(msg, trimFn)
						l = slog.LevelInfo
					case debug:
						msg = s
					default:
						continue
					}
					slog.Log(context.Background(), l, "["+app+"] "+msg)
				}
				return
			}
		}

		if ce := ParallelCall(ctx, readLog(spr, name+".1", debug), readLog(epr, name+".2", debug)); ce != nil {
			slog.Warn("readlog done", "err", ce)
		} else {
			slog.Info("readlog done")
		}
	}

	if err = c.Wait(); err != nil {
		if c.ProcessState != nil {
			if status, ok := c.ProcessState.Sys().(syscall.WaitStatus); ok {
				if status.Exited() && status.Signaled() {
					slog.Info(name+" exited", "signal", status.Signal().String())
					return
				}
			}
		}
		err = fmt.Errorf(name+" exited: %w", err)
	}
	return
}
