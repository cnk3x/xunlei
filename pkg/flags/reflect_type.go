package flags

import (
	"reflect"
	"strings"
)

func rType(t reflect.Type, noExtend ...bool) string {
	if len(noExtend) == 0 || !noExtend[0] {
		if te := GetExtend(t); te != nil {
			return te.Type()
		}
	}

	switch kind := t.Kind(); kind {
	case reflect.Pointer:
		return rType(t.Elem(), noExtend...)
	case reflect.Slice:
		return rType(t.Elem(), noExtend...) + "s"
	default:
		s := t.String()
		for i := len(s) - 1; i >= 0 && s[i] != '/'; i-- {
			if s[i] == '.' {
				return strings.ToLower(s[i+1:])
			}
		}
		return strings.ToLower(s)
	}
}

// 判断类型是否基础类型: int*, uint*, float*, string, bool
func isBasic(t reflect.Type) bool { return isBasicKind(t.Kind()) }

func isKnown(t reflect.Type) bool { return IsExtend(t) || isBasic(t) }

func isAllow(in reflect.Type) bool {
	var checkTypeInternal func(t reflect.Type, p, s bool) bool
	checkTypeInternal = func(t reflect.Type, p, s bool) bool {
		if isKnown(t) {
			return true
		}
		switch t.Kind() {
		case reflect.Pointer:
			return p && checkTypeInternal(t.Elem(), false, s)
		case reflect.Slice:
			return (s && checkTypeInternal(t.Elem(), true, false))
		default:
			return false
		}
	}

	return isKnown(in) || checkTypeInternal(in, true, true)
}

func bits(t reflect.Type) (bitsize int) {
	if isNumKind(t.Kind()) {
		bitsize = t.Bits()
	}
	return
}

func vBits(v reflect.Value) (bitsize int) { return bits(v.Type()) }

func isBasicKind(kind reflect.Kind) bool {
	return kind == reflect.String || kind == reflect.Bool || isNumKind(kind)
}

func isNumKind(kind reflect.Kind) bool {
	return isIntKind(kind) || isUintKind(kind) || isFloatKind(kind)
}

func isIntKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func isUintKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func isFloatKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func isStructType(t reflect.Type) bool { return typeIndirect(t).Kind() == reflect.Struct }

func typeIndirect(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}
