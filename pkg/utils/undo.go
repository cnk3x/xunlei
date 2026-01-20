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
