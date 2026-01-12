package utils

import (
	"reflect"
	"time"
)

// SelectOnce select `cSelect` chan, return it
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
//		 case v, ok:= cSeelct:
//			v = v, ok = ok, breaked = false
//	}
func SelectOnce[T any](cSelect <-chan T, cBreaks ...<-chan struct{}) (v T, ok, breaked bool) {
	c2case := func(c <-chan struct{}, _ int) reflect.SelectCase {
		return reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c)}
	}
	cases := append(Map(cBreaks, c2case), reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(cSelect)})
	if chosen, x, ok := reflect.Select(cases); chosen == len(cases)-1 {
		v, _ = x.Interface().(T)
		return v, ok, false
	}
	return v, false, true
}

// SelectDo do if recv from sSelect
func SelectDo[T any](cSelect <-chan T, okFunc func(), cBreaks ...<-chan struct{}) {
	if _, _, breaked := SelectOnce(cSelect, cBreaks...); !breaked {
		okFunc()
	}
}

func Sleep(d time.Duration, cBreaks ...<-chan struct{}) (breaked bool) {
	t := time.NewTimer(d)
	defer t.Stop()
	_, _, breaked = SelectOnce(t.C, cBreaks...)
	return
}
