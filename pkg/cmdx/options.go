package cmdx

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"syscall"

	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/cnk3x/xunlei/pkg/utils"
)

type Option func(*Cmd) (Undo, error)

func Options(options ...Option) Option {
	return func(c *Cmd) (undo Undo, err error) {
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
	attrInit := May(func(c *Cmd) {
		if c.SysProcAttr == nil {
			c.SysProcAttr = &syscall.SysProcAttr{}
		}
	})
	return Options(slices.Insert(options, 0, attrInit)...)
}

func Credential(uid, gid uint32) Option {
	var currentUid = os.Getuid()
	return ProcAttr(May(func(c *Cmd) error {
		if uid != 0 || gid != 0 {
			if currentUid != 0 {
				return fmt.Errorf("%s: root required, current uid: %d", os.ErrPermission, currentUid)
			}
			c.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, NoSetGroups: true}
		}
		return nil
	}))
}

func Args(args ...string) Option {
	return May(func(c *Cmd) { c.Args = append(c.Args[:1], args...) })
}

const True = "___true"

func Flags(args ...string) Option {
	var selectedArgs []string
	for i := 0; i < len(args)-1; i += 2 {
		if k, v := args[i], args[i+1]; k != "" && v != "" {
			if v == True {
				selectedArgs = append(selectedArgs, k)
			} else {
				selectedArgs = append(selectedArgs, k, v)
			}
		}
	}
	return Args(selectedArgs...)
}

func Dir(dir string) Option               { return May(func(c *Cmd) { c.Dir = dir }) }
func Env(env []string) Option             { return May(func(c *Cmd) { c.Env = env }) }
func Stderr(w io.Writer) Option           { return May(func(c *Cmd) { c.Stderr = w }) }
func Stdout(w io.Writer) Option           { return May(func(c *Cmd) { c.Stdout = w }) }
func Stdin(r io.Reader) Option            { return May(func(c *Cmd) { c.Stdin = r }) }
func PreStart(f func(*Cmd) error) Option  { return May(func(c *Cmd) { c.preStart = f }) }
func OnStarted(f func(*Cmd) error) Option { return May(func(c *Cmd) { c.onStarted = f }) }
func OnExit(f func(*Cmd) error) Option    { return May(func(c *Cmd) { c.onExit = f }) }

func LogStd() Option {
	return May(func(c *Cmd) { c.Stderr, c.Stdout = os.Stderr, os.Stdout })
}

func LineOut(lineRecv func(string)) Option {
	return May(func(c *Cmd) Undo {
		w := LineWriter(lineRecv)
		c.Stdout, c.Stderr = w, w
		return eUndo(w.Close)
	})
}

func LineErr(lineRecv func(string)) Option {
	return May(func(c *Cmd) Undo {
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

func May[O IOption](option O) func(*Cmd) (undo Undo, err error) {
	return func(c *Cmd) (undo Undo, err error) {
		return apply(option, c)
	}
}

type Undo = func()

func eUndo[E any](f func() E) Undo { return func() { _ = f() } }

func apply[MO IOption](option MO, t *Cmd) (undo Undo, err error) {
	undo = func() {}

	if optApply, ok := any(option).(func(*Cmd)); ok {
		optApply(t)
		return
	}

	if optApply, ok := any(option).(func(*Cmd) error); ok {
		err = optApply(t)
		return
	}

	if optApply, ok := any(option).(func(*Cmd) Undo); ok {
		undo = optApply(t)
		return
	}

	if optApply, ok := any(option).(func(*Cmd) (Undo, error)); ok {
		undo, err = optApply(t)
		return
	}

	err = fmt.Errorf("[%T]%v: apply error", option, option)
	return
}

type IOption interface {
	~func(*Cmd) (Undo, error) | ~func(*Cmd) Undo | ~func(*Cmd) error | ~func(*Cmd)
}
