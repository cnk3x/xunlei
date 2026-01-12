package flags

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/spf13/pflag"
)

func init() {
	pflag.ErrHelp = fmt.Errorf("\nstart with %s [...OPTIONS]", filepath.Base(os.Args[0]))
	pflag.Usage = func() {
		fmt.Fprint(os.Stderr, filepath.Base(os.Args[0]))
		if Version != "" {
			fmt.Fprint(os.Stderr, " - version "+Version)
		}
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "wrap system env as synology for xunlei")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "OPTIONS:")
		pflag.PrintDefaults()
	}
}

var (
	CommandLine = pflag.CommandLine
	Version     string
)

type FlagSet = pflag.FlagSet

func Var[T any](v T, name, short, usage string, env ...string) {
	VarFlag(CommandLine, v, name, short, usage, env...)
}

func Parse() (err error) { return ParseFlag(CommandLine) }

func VarFlag[T any](fSet *pflag.FlagSet, v T, name, short, usage string, env ...string) {
	var s string
	if len(env) > 0 {
		s = GetEnv("", env...)
		usage = fmt.Sprintf("%s%s[%s]", usage, utils.Iif(usage != "", " ", ""), strings.Join(env, ", "))
	}

	switch x := any(v).(type) {
	case *time.Duration:
		fSet.DurationVarP(x, name, short, *x, usage)
	case *net.IP:
		fSet.IPVarP(x, name, short, *x, usage)
	case *net.IPNet:
		fSet.IPNetVarP(x, name, short, *x, usage)
	case *string:
		fSet.StringVarP(x, name, short, *x, usage)
	case *int:
		fSet.IntVarP(x, name, short, *x, usage)
	case *int8:
		fSet.Int8VarP(x, name, short, *x, usage)
	case *int16:
		fSet.Int16VarP(x, name, short, *x, usage)
	case *int32:
		fSet.Int32VarP(x, name, short, *x, usage)
	case *int64:
		fSet.Int64VarP(x, name, short, *x, usage)
	case *uint:
		fSet.UintVarP(x, name, short, *x, usage)
	case *uint8:
		fSet.Uint8VarP(x, name, short, *x, usage)
	case *uint16:
		fSet.Uint16VarP(x, name, short, *x, usage)
	case *uint32:
		fSet.Uint32VarP(x, name, short, *x, usage)
	case *uint64:
		fSet.Uint64VarP(x, name, short, *x, usage)
	case *float32:
		fSet.Float32VarP(x, name, short, *x, usage)
	case *float64:
		fSet.Float64VarP(x, name, short, *x, usage)
	case *bool:
		fSet.BoolVarP(x, name, short, *x, usage)
	case *[]time.Duration:
		fSet.DurationSliceVarP(x, name, short, *x, usage)
	case *[]net.IP:
		fSet.IPSliceVarP(x, name, short, *x, usage)
	case *[]net.IPNet:
		fSet.IPNetSliceVarP(x, name, short, *x, usage)
	case *[]string:
		fSet.StringSliceVarP(x, name, short, *x, usage)
	case *[]int:
		fSet.IntSliceVarP(x, name, short, *x, usage)
	case *[]int32:
		fSet.Int32SliceVarP(x, name, short, *x, usage)
	case *[]int64:
		fSet.Int64SliceVarP(x, name, short, *x, usage)
	case *[]uint:
		fSet.UintSliceVarP(x, name, short, *x, usage)
	case *[]float32:
		fSet.Float32SliceVarP(x, name, short, *x, usage)
	case *[]float64:
		fSet.Float64SliceVarP(x, name, short, *x, usage)
	case *[]bool:
		fSet.BoolSliceVarP(x, name, short, *x, usage)
	default:
		panic(fmt.Errorf("%s type %v(%T) not support", name, x, x))
	}

	if s != "" {
		fSet.Set(name, s)
	}
}

func ParseFlag(fSet *pflag.FlagSet) (err error) {
	fSet.SortFlags = false
	return fSet.Parse(os.Args[1:])
}

func GetEnv(def string, keys ...string) (s string) {
	if len(keys) > 0 {
		for _, e := range keys {
			if s = os.Getenv(e); s != "" {
				return
			}
		}
	}
	return def
}
