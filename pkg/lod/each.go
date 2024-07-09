package lod

import "slices"

// Map returns a new slice of the passed in slice, applying the passed in function to each item.
func Map[I any, O any, F func(I) O | func(I, int) O | func(I) (O, bool) | func(I, int) (O, bool)](s []I, f F) (r []O) {
	if f == nil {
		return
	}

	var process func(I, int) (O, bool)
	if f0, ok := any(f).(func(I, int) (O, bool)); ok {
		process = f0
	} else if f1, ok := any(f).(func(I) O); ok {
		process = func(v I, i int) (O, bool) { return f1(v), true }
	} else if f2, ok := any(f).(func(I, int) O); ok {
		process = func(v I, i int) (O, bool) { return f2(v, i), true }
	} else if f3, ok := any(f).(func(I) (O, bool)); ok {
		process = func(v I, i int) (O, bool) { return f3(v) }
	}

	r = make([]O, 0, len(s))
	for i, v := range s {
		item, ok := process(v, i)
		if !ok {
			break
		}
		r = append(r, item)
	}
	return r
}

// Flatten returns a new slice concatenating the passed in slices.
func Flatten[T any](s [][]T) []T { return slices.Concat(s...) }

// ForEach iterates over the passed in slices and calls the passed in function.
func ForEach[I any, F func(I) | func(I, int) | func(I) bool | func(I, int) bool](s []I, f F) {
	if f == nil {
		return
	}

	var process func(I, int) bool
	if f0, ok := any(f).(func(I, int) bool); ok {
		process = f0
	} else if f1, ok := any(f).(func(I) bool); ok {
		process = func(v I, i int) bool { return f1(v) }
	} else if f2, ok := any(f).(func(I, int)); ok {
		process = func(v I, i int) bool { f2(v, i); return true }
	} else if f3, ok := any(f).(func(I)); ok {
		process = func(v I, i int) bool { f3(v); return true }
	}

	for i, v := range s {
		if !process(v, i) {
			break
		}
	}
}

// Filter returns a new slice of the passed in slice, applying the passed in function to each item.
func Filter[S ~[]I, I any, F func(I) bool | func(I, int) bool](s S, f F) (r S) {
	if f == nil {
		return s
	}

	var predicate func(I, int) bool
	if f0, ok := any(f).(func(I, int) bool); ok {
		predicate = f0
	} else if f1, ok := any(f).(func(I) bool); ok {
		predicate = func(v I, i int) bool { return f1(v) }
	}

	if predicate == nil {
		return s
	}

	r = make(S, 0, len(s))
	for i, v := range s {
		if predicate(v, i) {
			r = append(r, v)
		}
	}
	return
}
