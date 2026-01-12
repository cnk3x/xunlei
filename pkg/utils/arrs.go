package utils

import "cmp"

func CompactUniq[Slice ~[]T, T comparable](s Slice, inplace ...bool) Slice {
	result := s
	if !cmp.Or(inplace...) {
		result = make(Slice, len(s))
	}

	seen := make(map[T]struct{}, len(s))
	var zero T
	var x int
	for _, v := range s {
		if v == zero {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result[x], x = v, x+1
	}
	return result[:x]
}

func Conv[Slice ~[]T, T any](s Slice, cleanFn func(T) T, inplace ...bool) Slice {
	result := s
	if !cmp.Or(inplace...) {
		result = make(Slice, len(s))
	}
	for i, it := range s {
		result[i] = cleanFn(it)
	}
	return result
}

func Map[T, R any](s []T, conv func(T, int) R) []R {
	result := make([]R, len(s))
	for i, v := range s {
		result[i] = conv(v, i)
	}
	return result
}

func Reduce[T, R any](s []T, walk func(agg R, item T, i int) R, init R) R {
	for i, v := range s {
		init = walk(init, v, i)
	}
	return init
}

func Flat[T any](s [][]T) []T {
	l := Reduce(s, func(agg int, item []T, _ int) int { return agg + len(item) }, 0)
	return Reduce(s, func(agg []T, item []T, _ int) []T { return append(agg, item...) }, make([]T, 0, l))
}
