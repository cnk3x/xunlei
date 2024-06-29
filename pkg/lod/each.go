package lod

import "slices"

// 集合映射
func Map[S any, R any](s []S, f func(S) R) []R {
	r := make([]R, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}

// 集合映射
func MapIndex[S any, R any](s []S, f func(S, int) R) []R {
	r := make([]R, len(s))
	for i, v := range s {
		r[i] = f(v, i)
	}
	return r
}

// Flatten returns a new slice concatenating the passed in slices.
func Flatten[T any](s [][]T) []T { return slices.Concat(s...) }
