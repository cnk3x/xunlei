package flags

import (
	"reflect"
)

var extends = extendMap{}

type ExtendType interface {
	// Get 用于根据传入的reflect.Value获取对应的字符串。
	//
	// 参数:
	//  - v - 代表要获取数据的reflect.Value。
	Get(v reflect.Value) (s string)

	// Set 用于通过反射设置值。
	//
	// 参数:
	//   - v: 要设置值的反射值。
	//   - s: 要设置的字符串值。
	Set(v reflect.Value, s string, reset bool) (err error)

	// Type 值的类型字符串。
	Type() string
}

// Extend 为指定类型 T 注册自定义的解析和格式化函数。
//
//	parse 函数用于将字符串解析为 T 类型的值，format 函数用于将 T 类型的值格式化为字符串。
//
// 参数:
//   - parse: 一个函数，其输入为字符串，输出为 T 类型的值和可能的错误。用于将配置文件中的字符串值解析为实际的类型 T。
//   - format: 一个函数，其输入为 T 类型的值，输出为字符串。用于将类型 T 的值格式化为字符串，以便写入配置文件中。
func Extend[T any](parse func(src T, s string) (T, error), format func(T) string, typeString string) {
	setFunc := func(v reflect.Value, s string, reset bool) (err error) {
		if parse != nil {
			var src T
			if !reset {
				src, _ = v.Interface().(T)
			}
			if r, e := parse(src, s); e == nil {
				v.Set(reflect.ValueOf(r))
			} else {
				err = e
			}
		}
		return
	}

	getFunc := func(v reflect.Value) (s string) {
		if format != nil {
			r, _ := v.Interface().(T)
			s = format(r)
		}
		return
	}

	var x T
	typ := reflect.TypeOf(x)

	if typeString == "" {
		typeString = rType(typ, true)
	}

	st := &simpleType{typ: typeString, setFunc: setFunc, getFunc: getFunc}

	extends[typ] = st
}

// GetExtend 通过反射类型获取对应的扩展类型
//
// 参数:
//   - t: 要获取扩展类型的反射类型
//
// 返回值:
//   - ExtendType: 与给定反射类型对应的扩展类型
func GetExtend(t reflect.Type) ExtendType { return extends[t] }

// IsExtend 函数用于判断指定类型是否为扩展类型。
//
// 参数:
//   - t reflect.Type - 需要判断类型的 reflect.Type 实例。
//
// 返回值:
//   - bool - 如果类型是扩展类型，则返回 true；否则返回 false。
func IsExtend(t reflect.Type) (yes bool) { _, yes = extends[t]; return }

type extendMap map[reflect.Type]ExtendType

type simpleType struct {
	typ     string
	setFunc func(src reflect.Value, s string, reset bool) error
	getFunc func(src reflect.Value) string
}

func (te *simpleType) Get(v reflect.Value) (s string) {
	if te != nil && te.getFunc != nil {
		s = te.getFunc(v)
	}
	return
}

func (te *simpleType) Set(v reflect.Value, s string, reset bool) (err error) {
	if te != nil && te.setFunc != nil {
		if err = te.setFunc(v, s, reset); err != nil {
			return
		}
	}
	return
}

func (te *simpleType) Type() string { return te.typ }
