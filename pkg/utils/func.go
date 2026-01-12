package utils

func Fig[I, T, R any](f func(T) R) func(T, I) R {
	return func(t1 T, _ I) R { return f(t1) }
}

func Fig2[I, T1, T2, R any](f func(T1, T2) R) func(T1, T2, I) R {
	return func(t1 T1, t2 T2, _ I) R { return f(t1, t2) }
}
