package utils

import (
	"time"
)

// SelectOnce select `cSelect` chan, return it (blocked)
//
//	v: if recv from cSelect
//	ok: if cSelect is not closed
//	breaked: if recv from cBreaks or any cBreaks is closed...
//
// like:
//
//	select{
//		 case <- cBreaks[...]:
//			v = nil, ok = false, breaked = true
//		 case v, ok:= cSelect:
//			v = v, ok = ok, breaked = false
//	}
func CSelect[T any](cSelect <-chan T, cBreaks ...<-chan struct{}) (v T, ok, breaked bool) {
	cBreaked := make(chan struct{})

	for _, c := range cBreaks {
		go func() {
			select {
			case <-cBreaked:
			case <-c:
				close(cBreaked)
			}
		}()
	}

	select {
	case <-cBreaked:
		breaked = true
	case v, ok = <-cSelect:
		close(cBreaked)
	}
	return
}

// SelectDo do if recv from sSelect (blocked)
func SelectDo[T any](cSelect <-chan T, f func(T, bool), cBreaks ...<-chan struct{}) (breaked bool) {
	cBreaked := make(chan struct{})

	for _, c := range cBreaks {
		go func() {
			select {
			case <-cBreaked:
			case <-c:
				close(cBreaked)
			}
		}()
	}

	select {
	case <-cBreaked:
		return true
	case v, ok := <-cSelect:
		close(cBreaked)
		f(v, ok)
		return false
	}
}

// Sleep breakable sleep (blocked)
func Sleep(d time.Duration, cBreaks ...<-chan struct{}) (breaked bool) {
	t := time.NewTimer(d)
	defer t.Stop()
	_, _, breaked = CSelect(t.C, cBreaks...)
	return
}

// After do when done is closed (unblocked)
func After[T any, F ft[T]](done <-chan T, f F, cBreaks ...<-chan struct{}) {
	go SelectDo(done, func(t T, ok bool) {
		af := any(f)
		if ft, ok := af.(func()); ok {
			ft()
			return
		}

		if ft, ok := af.(func() error); ok {
			_ = ft()
			return
		}

		if ft, ok := af.(func(T)); ok {
			ft(t)
			return
		}

		if ft, ok := af.(func(T) error); ok {
			_ = ft(t)
			return
		}
	}, cBreaks...)
}

type ft[T any] interface {
	~func() | ~func() error | ~func(T) | ~func(T) error
}
