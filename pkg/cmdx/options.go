package cmdx

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/utils"
)

type Option func(*exec.Cmd) (Undo, error)

func Options(options ...Option) Option {
	return func(c *exec.Cmd) (undo Undo, err error) {
		pUndo := utils.MakeUndoPool(&undo, &err)
		defer pUndo.ErrDefer()
		for _, option := range options {
			var closer Undo
			if closer, err = apply(option, c); err != nil {
				return
			}
			pUndo.Put(closer)
		}
		return
	}
}

func ProcAttr(options ...Option) Option {
	attrInit := May(func(c *exec.Cmd) {
		if c.SysProcAttr == nil {
			c.SysProcAttr = &syscall.SysProcAttr{}
		}
	})
	return Options(slices.Insert(options, 0, attrInit)...)
}

func Credential(uid, gid uint32) Option {
	return ProcAttr(May(func(c *exec.Cmd) error {
		if uid != 0 || gid != 0 {
			if !IsRoot() {
				return fmt.Errorf("%s: root required, current uid: %d", os.ErrPermission, currentUid)
			}
			c.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, NoSetGroups: true}
		}
		return nil
	}))
}

func Args(args ...string) Option {
	return May(func(c *exec.Cmd) { c.Args = append(c.Args[:1], args...) })
}

func Dir(dir string) Option     { return May(func(c *exec.Cmd) { c.Dir = dir }) }
func Env(env []string) Option   { return May(func(c *exec.Cmd) { c.Env = env }) }
func Stderr(w io.Writer) Option { return May(func(c *exec.Cmd) { c.Stderr = w }) }
func Stdout(w io.Writer) Option { return May(func(c *exec.Cmd) { c.Stdout = w }) }
func Stdin(r io.Reader) Option  { return May(func(c *exec.Cmd) { c.Stdin = r }) }

func LogStd() Option {
	return May(func(c *exec.Cmd) { c.Stderr, c.Stdout = os.Stderr, os.Stdout })
}

func LineOut(lineRecv func(string)) Option {
	return May(func(c *exec.Cmd) Undo {
		w := LineWriter(lineRecv)
		c.Stdout, c.Stderr = w, w
		return eUndo(w.Close)
	})
}

func LineErr(lineRecv func(string)) Option {
	return May(func(c *exec.Cmd) Undo {
		w := LineWriter(lineRecv)
		c.Stderr = w
		return eUndo(w.Close)
	})
}

func LineIn(lines ...string) Option {
	return Stdin(strings.NewReader(strings.Join(lines, "\n") + "\n"))
}

func SlogDebug(prefix string) Option {
	le := LineErr(func(s string) { slog.Debug("[stderr] "+s, log.PrefixAttr(prefix)) })
	lo := LineOut(func(s string) { slog.Debug("[stdout] "+s, log.PrefixAttr(prefix)) })
	return Options(le, lo)
}

func May[O IOption](option O) func(*exec.Cmd) (undo Undo, err error) {
	return func(c *exec.Cmd) (undo Undo, err error) {
		return apply(option, c)
	}
}

type Undo = func()

func eUndo[E any](f func() E) Undo { return func() { _ = f() } }

func apply[MO IOption](option MO, t *exec.Cmd) (undo Undo, err error) {
	if apply, ok := any(option).(func(*exec.Cmd)); ok {
		apply(t)
		return
	}

	if apply, ok := any(option).(func(*exec.Cmd) error); ok {
		err = apply(t)
		return
	}

	if apply, ok := any(option).(func(*exec.Cmd) Undo); ok {
		undo = apply(t)
		return
	}

	if apply, ok := any(option).(func(*exec.Cmd) (Undo, error)); ok {
		undo, err = apply(t)
		return
	}

	return
}

type IOption interface {
	~func(*exec.Cmd) (Undo, error) | ~func(*exec.Cmd) Undo | ~func(*exec.Cmd) error | ~func(*exec.Cmd)
}
