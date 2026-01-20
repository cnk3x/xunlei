package authenticate_cgi

import (
	"os"
)

func WriteTo(w *os.File) error {
	if _, err := w.Write(Bytes); err != nil {
		return err
	}
	return w.Chmod(0777)
}
