package flags

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/spf13/pflag"
)

type pFlagSet = pflag.FlagSet

func New(name string) *FlagSet {
	f := &FlagSet{}
	f.Init(name)
	return f
}

type FlagSet struct {
	*pFlagSet
	Version           string
	Prefix            childPrefix
	HandleVersionFlag func(version string)

	errs []error
}

func (f *FlagSet) Init(name string) {
	out := os.Stderr

	if f.pFlagSet == nil {
		f.pFlagSet = pflag.NewFlagSet(name, pflag.ContinueOnError)
	}

	pflag.ErrHelp = fmt.Errorf("use %s [...Flags] to execute", name)

	f.SetOutput(out)
	f.SortFlags = false

	f.Usage = func() {
		fmt.Fprintf(out, "%s", name)
		if f.Version != "" {
			fmt.Fprintf(out, " -- version %s", f.Version)
		}
		fmt.Fprintf(out, "\n\n")
		fmt.Fprintf(out, "USAGE:\n")
		fmt.Fprintf(out, "  %s [...Flags]\n\n", name)
		fmt.Fprintf(out, "Flags:\n")
		fmt.Fprintln(out, f.FlagUsagesWrapped(0))
	}
}

func (f *FlagSet) SetVersion(ver string) { f.Version = sels(ver, f.Version) }

// Struct 是一个解析结构体的函数。
//
//	它接受一个指向结构体的指针作为参数，尝试使用默认配置解析传入的结构体。
//	如果解析失败，将错误信息打印到标准错误输出，并退出程序。
//
// 参数:
//   - structPtr any - 指向要解析的结构体的指针。
func (f *FlagSet) Struct(structPtr any, prefix *childPrefix) {
	if err := structBind(f, structPtr, prefix); err != nil {
		f.errs = append(f.errs, err)
	}
}

func (f *FlagSet) Var(obj any, name, shorthand, usage string) {
	v := newValue(Ref(obj))
	it := f.VarPF(v, name, shorthand, usage)
	if v.IsKind(reflect.Bool) {
		it.NoOptDefVal = "true"
	}
}

// ParseFlag 解析命令行参数。
//
// 参数:
//   - args []string: 命令行参数数组。
func (f *FlagSet) Parse(args []string) (err error) {
	if f.errs != nil {
		return errors.Join(f.errs...)
	}

	const versionFlag = "version"
	if f.Version != "" && f.Lookup(versionFlag) == nil {
		var shorthand string
		if f.Lookup("v") == nil {
			shorthand = "v"
		}
		if shorthand == "" && f.Lookup("V") == nil {
			shorthand = "V"
		}
		f.BoolP(versionFlag, shorthand, false, "显示版本信息")
	}

	if err = f.pFlagSet.Parse(args); err != nil {
		return
	}

	if ver, _ := f.GetBool(versionFlag); ver {
		if f.HandleVersionFlag != nil {
			f.HandleVersionFlag(f.Version)
		} else {
			fmt.Fprintf(os.Stdout, f.Version)
			os.Exit(0)
		}
	}

	return
}

var prefixDefault = &childPrefix{}

type childPrefix struct {
	Flag string
	Env  string
}
