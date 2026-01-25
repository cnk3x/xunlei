package utils

import "slices"

type funcQueue struct {
	Put      func(undo func())
	ErrDefer func()
	Run      func()
}

// BackQueue 先进后出的回滚函数队列
//   - rollback: 这是返回参数，回滚方法。
//   - autoRunIfErr: 这是传入参数，在 `defer fq.ErrDefer()` 中判断*error不为空则自动执行回滚
func BackQueue(rollback *func(), autoRunIfErr *error) (q funcQueue) {
	var funcs []func()

	q.Run = func() {
		for _, f := range slices.Backward(funcs) {
			if f != nil {
				f()
			}
		}
	}

	q.Put = func(f func()) {
		if f != nil {
			funcs = append(funcs, f)
		}
	}

	q.ErrDefer = func() {
		if autoRunIfErr != nil && *autoRunIfErr != nil {
			q.Run()
		}
	}

	if rollback != nil {
		*rollback = q.Run
	}
	return
}

func Run(fs ...func()) {
	for _, f := range fs {
		if f != nil {
			f()
		}
	}
}

func BackwardRun(fs ...func()) {
	for _, f := range slices.Backward(fs) {
		if f != nil {
			f()
		}
	}
}

func QRun(qr ...func() (undo func(), err error)) (undo func(), err error) {
	bq := BackQueue(&undo, &err)
	defer bq.ErrDefer()
	for _, q := range qr {
		if q != nil {
			u, e := q()
			if err = e; err != nil {
				return
			}
			bq.Put(u)
		}
	}
	return
}

func ERun(fs ...func() error) error {
	for _, f := range fs {
		if f != nil {
			if err := f(); err != nil {
				return err
			}
		}
	}
	return nil
}

func Fue1[T1 any](srcFn func(T1) (func(), error), t1 T1) func() (func(), error) {
	return func() (func(), error) { return srcFn(t1) }
}

func Fue2[T1, T2 any](srcFn func(T1, T2) (func(), error), t1 T1, t2 T2) func() (func(), error) {
	return func() (func(), error) { return srcFn(t1, t2) }
}

func Fue3[T1, T2, T3 any](srcFn func(T1, T2, T3) (func(), error), t1 T1, t2 T2, t3 T3) func() (func(), error) {
	return func() (func(), error) { return srcFn(t1, t2, t3) }
}

func Fue4[T1, T2, T3, T4 any](srcFn func(T1, T2, T3, T4) (func(), error), t1 T1, t2 T2, t3 T3, t4 T4) func() (func(), error) {
	return func() (func(), error) { return srcFn(t1, t2, t3, t4) }
}

func Fue4r[T1, T2, T3, T4 any](srcFn func(T1, T2, T3, ...T4) (func(), error), t1 T1, t2 T2, t3 T3, t4 ...T4) func() (func(), error) {
	return func() (func(), error) { return srcFn(t1, t2, t3, t4...) }
}

func Fe1[T1 any](srcFn func(T1) error, t1 T1) func() error {
	return func() error { return srcFn(t1) }
}

func Fnu(fn func() error) func() (func(), error) {
	return func() (func(), error) {
		return nil, fn()
	}
}
