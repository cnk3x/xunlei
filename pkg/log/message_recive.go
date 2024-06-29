package log

import (
	"bufio"
	"context"
	"io"
	"log"
)

func MessageRecive(ctx context.Context, handleMessage func(ctx context.Context, msg string)) io.WriteCloser {
	lw := &reciver{}
	lw.ctx, lw.cancel = context.WithCancel(ctx)
	lw.pr, lw.pw = io.Pipe()
	lw.done = make(chan struct{})
	lw.handleMessage = handleMessage

	go func() {
		defer close(lw.done)
		defer lw.pr.Close()
		defer lw.pw.Close()

		select {
		case <-ctx.Done():
			return
		default:
			lw.Scan()
		}
	}()

	return lw
}

type reciver struct {
	pr            io.ReadCloser
	pw            io.WriteCloser
	ctx           context.Context
	cancel        context.CancelFunc
	handleMessage func(ctx context.Context, msg string)
	done          chan struct{}
}

func (lw *reciver) Scan() {
	for br := bufio.NewScanner(lw.pr); br.Scan(); {
		select {
		case <-lw.ctx.Done():
			return
		default:
			lw.handleMessage(lw.ctx, br.Text())
		}
	}
}

func (lw *reciver) Write(p []byte) (n int, err error) {
	if n, err = lw.pw.Write(p); err != nil {
		lw.cancel()
	}
	return
}

func (lw *reciver) Close() (err error) {
	if lw.cancel != nil {
		lw.cancel()
	}
	<-lw.done
	return
}

func LogStd(w io.Writer, prefix string) *log.Logger { return log.New(w, prefix, 0) }
