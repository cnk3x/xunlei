package flags

import (
	"fmt"
	"reflect"
	"strconv"
)

func Ref(src any) reflect.Value {
	if r, ok := src.(reflect.Value); ok {
		return reflect.Indirect(r)
	}

	if v, ok := src.(*Value); ok {
		return reflect.Indirect(v.Ref)
	}

	return reflect.Indirect(reflect.ValueOf(src))
}

func rGet(v reflect.Value) (out []string) {
	if IsZero(v) {
		return
	}

	if te := GetExtend(v.Type()); te != nil {
		return o2s(te.Get(v))
	}

	switch kind := v.Kind(); kind {
	case reflect.Pointer:
		out = rGet(v.Elem())
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			out = append(out, rGet(v.Index(i))...)
		}
	case reflect.String:
		out = o2s(v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out = o2s(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		out = o2s(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		out = o2s(strconv.FormatFloat(v.Float(), 'f', -1, 64))
	case reflect.Bool:
		out = o2s(strconv.FormatBool(v.Bool()))
	}
	return
}

func rSet(v reflect.Value, s string, reset bool) (err error) {
	if !v.IsValid() {
		return invalid("rSet")
	}

	if !v.CanSet() && v.Kind() != reflect.Pointer {
		return fmt.Errorf("value can not set, kind=%s", v.Kind())
	}

	if te := GetExtend(v.Type()); te != nil {
		return te.Set(v, s, reset)
	}

	switch kind := v.Kind(); kind {
	case reflect.String:
		v.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if r, e := strconv.ParseInt(s, 0, vBits(v)); e == nil {
			v.SetInt(r)
		} else {
			err = e
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if r, e := strconv.ParseUint(s, 0, vBits(v)); e == nil {
			v.SetUint(r)
		} else {
			err = e
		}
	case reflect.Float32, reflect.Float64:
		if r, e := strconv.ParseFloat(s, vBits(v)); e == nil {
			v.SetFloat(r)
		} else {
			err = e
		}
	case reflect.Bool:
		if r, e := strconv.ParseBool(s); e == nil {
			v.SetBool(r)
		} else {
			err = e
		}
	case reflect.Pointer:
		err = rSet(mkPtr(v).Elem(), s, reset)
	case reflect.Slice:
		if reset {
			v.Set(v.Slice(0, 0))
		}

		var el reflect.Value
		if et := v.Type().Elem(); et.Kind() == reflect.Pointer {
			el = reflect.New(et.Elem())
		} else {
			el = reflect.New(et).Elem()
		}

		if err = rSet(el, s, false); err != nil {
			return
		}

		v.Set(reflect.Append(v, el))
	default:
		err = fmt.Errorf("unknown kind: %s", kind)
	}

	return
}

// IsZero reports whether v is the zero value for its type.
//
//	It return true if the argument is invalid.
func IsZero(v reflect.Value) bool {
	kind := v.Kind()
	return kind == reflect.Invalid || v.IsZero()
}

// IsNil reports whether its argument v is nil.
//
//	The argument must be a chan, func, interface, map, pointer or slice value, if it is not, return it is invalid.
func IsNil(v reflect.Value) bool {
	kind := v.Kind()
	switch {
	case kind == reflect.Invalid:
		return true
	case kind >= reflect.Chan && kind <= reflect.Slice:
		return v.IsNil()
	case kind == reflect.UnsafePointer:
		return v.IsNil()
	default:
		return false
	}
}
