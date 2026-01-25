package log

import (
	"bufio"
	"io"
)

func Line(lineRecv func(s string)) io.WriteCloser {
	r, w := io.Pipe()
	go func() {
		for scan := bufio.NewScanner(r); scan.Scan(); {
			lineRecv(scan.Text())
		}
	}()
	return w
}
