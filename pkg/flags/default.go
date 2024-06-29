package flags

import (
	"fmt"
	"os"
	"path/filepath"
)

var Default = New(filepath.Base(os.Args[0]))

func SetVersion(ver string) {
	Default.SetVersion(ver)
}

func Struct(obj any) {
	Default.Struct(obj, prefixDefault)
}

func Var(obj any, name, shorthand, usage string) {
	Default.Var(obj, name, shorthand, usage)
}

func Parse() {
	if err := Default.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}
}
