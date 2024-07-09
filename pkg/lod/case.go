package lod

// Iif 三元运算
func Iif[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}

// IifF 三元运算
func IifF[T any](cond bool, t, f func() T) (out T) {
	if r := Iif(cond, t, f); r != nil {
		out = r()
	}
	return
}

func If[T any](c bool, v T) IfElse[T]         { return Iif(c, v2f(v), nil) }
func IfF[T any](c bool, v func() T) IfElse[T] { return Iif(c, v, nil) }

type IfElse[T any] func() T

func (i IfElse[T]) ElseIfF(c bool, v func() T) IfElse[T] { return Iif(c, v, i) }
func (i IfElse[T]) ElseF(v func() T) T                   { return IifF(i == nil, v, i) }
func (i IfElse[T]) ElseIf(c bool, v T) IfElse[T]         { return i.ElseIfF(c, v2f(v)) }
func (i IfElse[T]) Else(v T) T                           { return i.ElseF(v2f(v)) }

func v2f[T any](v T) func() T { return func() T { return v } }
