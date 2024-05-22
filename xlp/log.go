package xlp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
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
			slog.Warn("readlog done")
		}
	}

	if err = c.Wait(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			slog.Error(fmt.Sprintf(name+" stopped: %s", string(ee.Stderr)))
		} else {
			err = fmt.Errorf(name+" stopped: %w", err)
		}
	}
	return
}
