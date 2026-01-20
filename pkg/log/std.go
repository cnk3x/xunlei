package log

import (
	"context"
	"io"
	"log/slog"
	"runtime"
	"time"
)

func Writer(ctx context.Context) io.Writer {
	return &stdWriter{ctx: ctx, h: slog.Default().Handler()}
}

func Caller(skip int) (pc uintptr) {
	var pcs [1]uintptr
	runtime.Callers(2+skip, pcs[:])
	return pcs[0]
}

func ErrDebug(err error) slog.Level {
	if err != nil {
		return slog.LevelError
	}
	return slog.LevelDebug
}

func WarnDebug(err error) slog.Level {
	if err != nil {
		return slog.LevelWarn
	}
	return slog.LevelDebug
}

type stdWriter struct {
	ctx       context.Context
	level     slog.Leveler
	h         slog.Handler
	capturePC bool
}

func (w *stdWriter) Write(buf []byte) (int, error) {
	level := w.level.Level()
	if !w.h.Enabled(w.ctx, level) {
		return 0, nil
	}
	var pc uintptr
	if w.capturePC {
		pc = Caller(3)
	}

	origLen := len(buf)
	if len(buf) > 0 && buf[len(buf)-1] == '\n' {
		buf = buf[:len(buf)-1]
	}
	r := slog.NewRecord(time.Now(), level, string(buf), pc)
	return origLen, w.h.Handle(w.ctx, r)
}
