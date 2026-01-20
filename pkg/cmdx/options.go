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

type Option = func(*Cmd) (Closer, error)

func Options(options ...Option) Option {
	return func(c *Cmd) (closer Closer, err error) {
		bq := utils.BackQueue(&closer, &err)
		defer bq.ErrDefer()
		for _, option := range options {
			fc, e := option(c)
			if err = e; err != nil {
				return
			}
			bq.Put(fc)
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

// Credential 以另外的身份运行外部程序
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

// Args 覆盖设置外部程序运行参数，替换掉 `exec.Command(name, args...)` 中所有的args
func Args(args ...string) Option {
	return May(func(c *Cmd) { c.Args = append(c.Args[:1], args...) })
}

const True = "___true"

// Flags 覆盖设置外部程序运行参数，以两个一对键值对分割
//   - 如果是无值参数，需添加[cmdx.True]来占位
//   - 如果提供的键或值为空，该方法会自动忽略。
//   - 有空值要求的，请使用 [cmdx.Args] 选项
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

func LineOut(lineRecv func(string)) Option {
	return May(func(c *Cmd) Closer {
		w := LineWriter(lineRecv)
		c.Stdout, c.Stderr = w, w
		return closerNe(w.Close)
	})
}

func LineErr(lineRecv func(string)) Option {
	return May(func(c *Cmd) Closer {
		w := LineWriter(lineRecv)
		c.Stderr = w
		return closerNe(w.Close)
	})
}

func LineIn(lines ...string) Option {
	return Stdin(strings.NewReader(strings.Join(lines, "\n") + "\n"))
}

func Slog(prefix string, level slog.Level) Option {
	le := LineErr(func(s string) { slog.Debug("[stderr] "+s, log.PrefixAttr(prefix)) })
	lo := LineOut(func(s string) { slog.Debug("[stdout] "+s, log.PrefixAttr(prefix)) })
	return Options(le, lo)
}

func LogStd() Option {
	return May(func(c *Cmd) { c.Stderr, c.Stdout = os.Stderr, os.Stdout })
}

type iOption interface {
	~func(*Cmd) | ~func(*Cmd) error | ~func(*Cmd) Closer | ~func(*Cmd) (Closer, error)
}

func May[O iOption](option O) func(*Cmd) (fc Closer, err error) {
	return func(c *Cmd) (Closer, error) {
		return apply(option, c)
	}
}

type Closer = func()

func closerNe[E any](f func() E) Closer { return func() { _ = f() } }

func apply[MO iOption](option MO, t *Cmd) (fc Closer, err error) {
	fc = func() {}

	if optApply, ok := any(option).(func(*Cmd) (Closer, error)); ok {
		fc, err = optApply(t)
		return
	}

	if optApply, ok := any(option).(func(*Cmd) Closer); ok {
		fc = optApply(t)
		return
	}

	if optApply, ok := any(option).(func(*Cmd) error); ok {
		err = optApply(t)
		return
	}

	if optApply, ok := any(option).(func(*Cmd)); ok {
		optApply(t)
		return
	}

	err = fmt.Errorf("[%T]%v: apply error", option, option)
	return
}
