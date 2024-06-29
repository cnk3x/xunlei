package lod

// Iif 三元运算
func Iif[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}

// Iif 三元运算
func IifF[T any](cond bool, t, f func() T) T { return Iif(cond, t, f)() }

// 选择第一个不为零值的值
func Select[T comparable](vs ...T) (out T) { return Selects(vs) }

// 选择第一个不为零值的值
func Selects[T comparable](vs []T) (out T) {
	for _, v := range vs {
		if v != out {
			return v
		}
	}
	return
}

func First[T any](vs []T) (out T) {
	if len(vs) > 0 {
		out = vs[0]
	}
	return
}

func Last[T any](vs []T) (out T) {
	if l := len(vs); l > 0 {
		out = vs[l-1]
	}
	return
}

func At[T any](vs []T, i int) (out T) {
	if 0 <= i && i < len(vs) {
		out = vs[i]
	}
	return
}
