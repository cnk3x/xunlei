package utils

import (
	"context"
	"errors"
	"sync"
)

func Fig[I, T, R any](f func(T) R) func(T, I) R {
	return func(t1 T, _ I) R { return f(t1) }
}

func Fig2[I, T1, T2, R any](f func(T1, T2) R) func(T1, T2, I) R {
	return func(t1 T1, t2 T2, _ I) R { return f(t1, t2) }
}

func FIdx[T, R any](f func(T) R) func(T, int) R { return Fig[int](f) }

func Fe(f func()) error            { f(); return nil }
func Fne[E any](f func() E) func() { return func() { _ = f() } }

func SeqExec(fns ...func() (err error)) (err error) {
	for _, fn := range fns {
		if err = fn(); err != nil {
			break
		}
	}
	return
}

func SeqExecWithUndo(fns ...func() (undo func(), err error)) (undo func(), err error) {
	pUndo := MakeUndoPool(&undo, &err)
	defer pUndo.ErrDefer()

	for _, fn := range fns {
		var u func()
		if u, err = fn(); err != nil {
			break
		}
		pUndo.Put(u)
	}

	return
}

var ErrDone = errors.New("done")

func SharedExec(fns ...func(ctx context.Context) (err error)) (err error) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(ErrDone)

	errs := make(chan error, len(fns))
	defer close(errs)

	wait := &sync.WaitGroup{}
	for _, fn := range fns {
		wait.Go(func() {
			if err := fn(ctx); err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					errs <- err
					cancel(err)
				}
			}
		})
	}
	wait.Wait()

	select {
	case <-ctx.Done():
	case err = <-errs:
	default:
	}

	return
}
