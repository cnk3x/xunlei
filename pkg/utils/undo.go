package utils

import "slices"

func MakeUndoPool(undo *func(), autoUndoIfErr *error) (r struct {
	Put      func(undo func())
	ErrDefer func()
	Run      func()
}) {
	var undos []func()

	r.Run = func() { BackwardCall(undos...) }

	r.Put = func(undo func()) {
		if undo != nil {
			undos = append(undos, undo)
		}
	}

	r.ErrDefer = func() {
		if autoUndoIfErr != nil && *autoUndoIfErr != nil {
			r.Run()
		}
	}

	if undo != nil {
		*undo = r.Run
	}
	return
}

func Call(fs ...func()) {
	for _, f := range fs {
		if f != nil {
			f()
		}
	}
}

func BackwardCall(fs ...func()) {
	for _, f := range slices.Backward(fs) {
		if f != nil {
			f()
		}
	}
}
