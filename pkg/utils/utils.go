package utils

import "strings"

func Iif[T any](c bool, t, f T) T {
	if c {
		return t
	}
	return f
}

func First[T any](v []T) T { return FirstOr(v) }
func FirstOr[T any](v []T, def ...T) T {
	if len(v) > 0 {
		return v[0]
	}
	if len(def) > 0 {
		return def[0]
	}
	var z T
	return z
}

func HasPrefix(s, prefix string, ignoreCase ...bool) bool {
	if First(ignoreCase) {
		return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
	}
	return strings.HasPrefix(s, prefix)
}

func Eol[T any](_ T, err error) error { return err }
func Eon[T any, E any](v T, _ E) T    { return v }
