package authenticate_cgi

import (
	"context"
	"os"

	"github.com/cnk3x/xunlei/pkg/fo"
)

func WriteTo(w *os.File) error {
	if _, err := w.Write(Bytes); err != nil {
		return err
	}
	return w.Chmod(0777)
}

func SaveFunc(ctx context.Context, fn string) func() (undo func(), err error) {
	return func() (undo func(), err error) { return fo.WriteFile(ctx, fn, WriteTo) }
}
