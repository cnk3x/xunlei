package utils

import (
	"bufio"
	"cmp"
	"io"
	"iter"
	"log"
)

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

func Conv[Slice ~[]T, T any](s Slice, cleanFn func(T) T, inplace ...bool) (result Slice) {
	result = s
	if !cmp.Or(inplace...) {
		result = make(Slice, len(s))
	}
	for i, it := range s {
		result[i] = cleanFn(it)
	}
	return
}

func Replace[Slice ~[]T, T any](s Slice, replaceFn func(T) (T, error)) (result Slice, err error) {
	result = make(Slice, len(s))
	for i, it := range s {
		if result[i], err = replaceFn(it); err != nil {
			result = result[:0]
			return
		}
	}
	return
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

func ReduceSeq2[T, R any](seq iter.Seq2[int, T], walk func(agg R, item T, index int) R, init R) R {
	for i, item := range seq {
		init = walk(init, item, i)
	}
	return init
}

func ReduceSeq[T, R any](seq iter.Seq[T], walk func(agg R, item T) R, init R) R {
	for item := range seq {
		init = walk(init, item)
	}
	return init
}

func LineSeq(r io.Reader) iter.Seq[string] {
	return func(yield func(string) bool) {
		for scan := bufio.NewScanner(r); scan.Scan(); {
			if !yield(scan.Text()) {
				break
			}
		}
	}
}

func LineWalk(r io.Reader, f func(s string)) {
	for scan := bufio.NewScanner(r); scan.Scan(); {
		f(scan.Text())
	}
}

func LineWriter(lineRead func(line string)) io.WriteCloser {
	r, w := io.Pipe()
	go LineWalk(r, lineRead)
	return w
}

func LogStd(w io.Writer) *log.Logger { return log.New(w, "", 0) }

func Array[T any](s ...T) []T { return s }
