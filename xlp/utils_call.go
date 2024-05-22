package xlp

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// 顺序执行
func GroupCall[F GroupCallT](ctx context.Context, calls ...F) (err error) {
	for _, call := range calls {
		if err = WrapCall(ctx, call)(); err != nil {
			break
		}
	}
	return
}

func ParallelCall(ctx context.Context, calls ...GroupCallFunc) error {
	g, c := errgroup.WithContext(ctx)
	for _, call := range calls {
		g.Go(WrapCall(c, call))
	}
	return g.Wait()
}

type GroupCallFunc func(ctx context.Context) (err error)

type GroupCallT interface {
	~func(ctx context.Context) (err error) | ~func() (err error)
}

func WrapCall[F GroupCallT](ctx context.Context, f F) func() (err error) {
	if call, ok := any(f).(GroupCallFunc); ok {
		return func() (err error) { return call(ctx) }
	}

	if call, ok := any(f).(func() (err error)); ok {
		return call
	}

	return func() (err error) { return fmt.Errorf("unsupport call type: %T", f) }
}
