package xlp

import "net/http"

// Iif 三元运算
func Iif[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}

// Iif 三元运算
func IifF[T any](cond bool, t, f func() T) T { return Iif(cond, t, f)() }

func Map[S any, R any](s []S, f func(S) R) []R {
	r := make([]R, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}

func Respond[T ~[]byte | ~string](body T, contentType string, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		w.Write([]byte(body))
	}
}

func SelectN[T comparable](v T, f ...T) (out T) {
	if v == out {
		for _, f := range f {
			if f != out {
				return f
			}
		}
	}
	return v
}
