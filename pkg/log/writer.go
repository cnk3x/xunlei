package log

import (
	"context"
	"io"
	"log/slog"
	"runtime"
	"time"
)

type Writer struct {
	h         slog.Handler
	level     slog.Leveler
	capturePC bool
	ctx       context.Context
}

func (w *Writer) Write(buf []byte) (int, error) {
	level := w.level.Level()
	if !w.h.Enabled(w.ctx, level) {
		return 0, nil
	}
	var pc uintptr
	if w.capturePC {
		// skip [w.Write, Logger.Output, log.Print]
		pc = Caller(3)
	}

	// Remove final newline.
	origLen := len(buf) // Report that the entire buf was written.
	if len(buf) > 0 && buf[len(buf)-1] == '\n' {
		buf = buf[:len(buf)-1]
	}
	r := slog.NewRecord(time.Now(), level, string(buf), pc)
	return origLen, w.h.Handle(w.ctx, r)
}

func NewWriter(ctx context.Context) io.Writer {
	return &Writer{h: slog.Default().Handler()}
}

func Caller(skip int) (pc uintptr) {
	// skip [runtime.Callers, log.Caller]
	var pcs [1]uintptr
	runtime.Callers(2+skip, pcs[:])
	pc = pcs[0]
	return
}
