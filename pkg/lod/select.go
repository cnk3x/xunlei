package lod

// Select 选择第一个不为零值的值
func Select[T comparable](vs ...T) (out T) { return Selects(vs) }

// Selects 选择第一个不为零值的值
func Selects[T comparable](vs []T, fallback ...T) (out T) {
	for _, v := range vs {
		if v != out {
			return v
		}
	}

	for _, v := range fallback {
		if v != out {
			return v
		}
	}
	return
}

// First 返回第一个值
func First[T any](vs []T, fallback ...T) (out T) {
	if len(vs) > 0 {
		return vs[0]
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return
}

// Last 返回最后一个值
func Last[T any](vs []T, fallback ...T) (out T) {
	if l := len(vs); l > 0 {
		return vs[l-1]
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return
}

// At 返回指定索引的值
func At[T any](vs []T, i int, fallback ...T) (out T) {
	if 0 <= i && i < len(vs) {
		return vs[i]
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return
}
