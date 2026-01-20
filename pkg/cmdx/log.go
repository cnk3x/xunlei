package cmdx

import (
	"bufio"
	"io"
)

func LineWriter(lineRead func(s string)) io.WriteCloser {
	r, w := io.Pipe()
	go func() {
		for scan := bufio.NewScanner(r); scan.Scan(); {
			lineRead(scan.Text())
		}
	}()
	return w
}
