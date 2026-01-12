package authenticate_cgi

import (
	"os"

	"github.com/cnk3x/xunlei/pkg/fo"
	"github.com/cnk3x/xunlei/pkg/utils"
)

func WriteTo(w *os.File) error {
	if err := utils.Eol(w.Write(Bytes)); err != nil {
		return err
	}
	return w.Chmod(0777)
}

func SaveTo(fn string) error {
	return fo.OpenWrite(fn, fo.Content(Bytes), fo.DirPerm(0777), fo.Perm(0777), fo.FlagExcl)
}
