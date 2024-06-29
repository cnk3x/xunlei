package flags

import "reflect"

func invalid(method string) error {
	return &reflect.ValueError{Method: method, Kind: reflect.Invalid}
}

func o2s[T any](n ...T) []T { return n }

func mkPtr(v reflect.Value) reflect.Value {
	if t := v.Type(); t.Kind() == reflect.Pointer && v.IsNil() {
		v.Set(reflect.New(t.Elem()))
	}
	return v
}

// sels 从数组中选择第一个不为零的值,参数是数组
func sel[T comparable](args []T) (out T) {
	for _, arg := range args {
		if arg != out {
			return arg
		}
	}
	return
}

// sels 从数组中选择第一个不为零的值,传入参数是变长参数
func sels[T comparable](args ...T) T { return sel(args) }

// selp 判断传入指针是否为零值，如果是则创建一个新值
func selp[T any](v *T, create func() *T) (out *T) {
	if out = v; out == nil {
		out = create()
	}
	return
}

var _ = selp[any]
