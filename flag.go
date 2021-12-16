package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
)

type Command struct {
	Name string
	Run  func(args []string)
	Desc string
}

func RunCommand(commands ...*Command) {
	var commandName string
	args := os.Args[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		commandName = args[0]
		args = args[1:]
	}

	usage := func() {
		maxN := 4
		for _, c := range commands {
			if l := len(c.Name); l > maxN {
				maxN = l
			}
		}
		prefix := os.Args[0]
		for _, c := range commands {
			fmt.Fprintf(os.Stderr, "%s %*s %s\n", prefix, -maxN, c.Name, c.Desc)
		}
		fmt.Fprintf(os.Stderr, "%s %*s %s\n\n", prefix, -maxN, "help", "显示帮助")
		os.Exit(1)
	}

	if commandName == "help" {
		usage()
	}

	for _, command := range commands {
		if commandName == command.Name {
			command.Run(args)
			return
		}
	}

	usage()
}

func Flag(n string, v interface{}, args []string) []string {
	type Flag struct {
		Name  string
		Short string
		ENV   string
		Usage string
		Value *Value
	}

	newFlag := func(fv reflect.Value, ft reflect.StructField) *Flag {
		name := ft.Tag.Get("flag")
		if name == "-" {
			return nil
		}
		if name == "" {
			name = strings.ToLower(ft.Name)
		}

		v := Flag{
			Value: &Value{Value: fv},
			Name:  name,
			Short: ft.Tag.Get("short"),
			ENV:   ft.Tag.Get("env"),
			Usage: ft.Tag.Get("usage"),
		}
		if v.ENV != "" {
			if s := os.Getenv(v.ENV); s != "" {
				v.Value.Set(s)
			}
		}
		return &v
	}

	flag.ErrHelp = errors.New("")
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.SetOutput(os.Stderr)

	var (
		rv         = reflect.Indirect(reflect.ValueOf(v))
		rt         = rv.Type()
		items      []*Flag
		nl, el, sl int
	)

	for i := 0; i < rt.NumField(); i++ {
		item := newFlag(rv.Field(i), rt.Field(i))
		if item == nil {
			continue
		}

		flagItem := flag.Flag{
			Name:      item.Name,
			Shorthand: item.Short,
			Usage:     item.Usage,
			Value:     item.Value,
			DefValue:  item.Value.String(),
		}
		if flagItem.Value.Type() == "bool" {
			flagItem.NoOptDefVal = "true"
		}
		fs.AddFlag(&flagItem)

		items = append(items, item)

		if sl == 0 && item.Short != "" {
			sl = 4
		}
		if l := len(item.Name) + 2; l > nl {
			nl = l
		}
		if l := len(item.ENV) + 3; l > el {
			el = l
		}
	}

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s [参数]\n", n)
		for _, item := range items {
			var ns, es string
			if item.Short != "" {
				ns = "-" + item.Short + ", --" + item.Name
			} else if sl > 0 {
				ns = "    --" + item.Name
			} else {
				ns = "--" + item.Name
			}

			if item.ENV != "" {
				es = " [" + item.ENV + "]"
			}
			def := item.Value.String()
			if def != "" {
				def = "(默认: " + def + ")"
			}
			fmt.Fprintf(os.Stderr, "  %*s%*s %s%s\n", -nl-sl, ns, -el, es, item.Usage, def)
		}
	}

	_ = fs.Parse(args)
	return fs.Args()
}

type Value struct {
	Value reflect.Value
}

func (v *Value) String() string {
	if arr := getString(v.Value); len(arr) > 0 {
		return arr[0]
	}
	return ""
}

func (v *Value) Set(s string) (err error) {
	return setString(v.Value, s)
}

func (v *Value) Type() string {
	return v.Value.Kind().String()
}

func (v *Value) Append(s string) (err error) {
	return v.Set(s)
}

func (v *Value) Replace(arr []string) error {
	v.Value.Set(reflect.Zero(v.Value.Type()))
	for _, s := range arr {
		if err := v.Set(s); err != nil {
			return err
		}
	}
	return nil
}

func (v *Value) GetSlice() (arr []string) {
	return getString(v.Value)
}

func setString(value reflect.Value, s string) error {
	switch value.Kind() {
	case reflect.Ptr:
		nv := reflect.New(value.Type().Elem())
		if err := setString(nv.Elem(), s); err != nil {
			return err
		}
		value.Set(nv)
	case reflect.String:
		value.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		value.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		value.SetUint(x)
	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		value.SetFloat(x)
	case reflect.Bool:
		x, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		value.SetBool(x)
	case reflect.Slice, reflect.Array:
		item := reflect.New(value.Type().Elem()).Elem()
		if err := setString(item, s); err != nil {
			return err
		}
		value.Set(reflect.Append(value, item))
	}
	return nil
}

func getString(value reflect.Value) (arr []string) {
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr,
		reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		if value.IsNil() {
			return
		}
	}

	switch value.Kind() {
	case reflect.Ptr:
		if value.IsZero() {
			return
		}
		return getString(value.Elem())
	case reflect.String:
		return []string{value.String()}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []string{strconv.FormatInt(value.Int(), 10)}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []string{strconv.FormatUint(value.Uint(), 10)}
	case reflect.Float32, reflect.Float64:
		return []string{strconv.FormatFloat(value.Float(), 'f', 2, 64)}
	case reflect.Bool:
		return []string{strconv.FormatBool(value.Bool())}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			arr = append(arr, getString(value.Index(i))...)
		}
		return
	default:
		return
	}
}
