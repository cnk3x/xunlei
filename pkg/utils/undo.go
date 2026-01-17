package utils

import "slices"

func MakeUndoPool(undo *func(), autoUndoIfErr *error) (r struct {
	Put      func(undo func())
	ErrDefer func()
	Run      func()
}) {
	var undos []func()

	r.Run = func() {
		for _, undo := range slices.Backward(undos) {
			if undo != nil {
				undo()
			}
		}
	}

	r.Put = func(undo func()) {}
	r.ErrDefer = func() {
		if autoUndoIfErr != nil && *autoUndoIfErr != nil {
			r.Run()
		}
	}

	*undo = r.Run
	return
}
