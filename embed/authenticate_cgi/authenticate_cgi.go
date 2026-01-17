package authenticate_cgi

import (
	"os"

	"github.com/cnk3x/xunlei/pkg/fo"
)

func WriteTo(w *os.File) error {
	if _, err := w.Write(Bytes); err != nil {
		return err
	}
	return w.Chmod(0777)
}

func SaveTo(fn string) error {
	return fo.OpenWrite(fn, WriteTo, fo.FlagExcl(true))
}
