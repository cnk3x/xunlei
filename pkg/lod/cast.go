package lod

import (
	"fmt"
	"reflect"
	"strconv"
)

// Itoa converts a number or bool to string
func Itoa[T Number | ~bool](in T) string {
	var z T
	if in == z {
		return ""
	}
	switch v := reflect.ValueOf(in); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		return "true"
	default:
		return fmt.Sprintf("%v", in)
	}
}

// Atoi converts a number or bool to string
func Atoi[T Number](in string) (out T, err error) {
	switch v := reflect.ValueOf(out); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var n int64
		n, err = strconv.ParseInt(in, 0, v.Type().Bits())
		out = T(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		var n uint64
		n, err = strconv.ParseUint(in, 0, v.Type().Bits())
		out = T(n)
	case reflect.Float32, reflect.Float64:
		var n float64
		n, err = strconv.ParseFloat(in, v.Type().Bits())
		out = T(n)
	default:
		err = fmt.Errorf("invalid number %q", in)
	}
	return
}

func FormatInt[T Int | Uint](in T, base ...int) string {
	return strconv.FormatInt(int64(in), First(base, 10))
}

var (
	ParseByte   = mkParseUint[byte](8)
	ParseUint8  = mkParseUint[uint8](8)
	ParseUint16 = mkParseUint[uint16](16)
	ParseUint32 = mkParseUint[uint32](32)
	ParseUint64 = mkParseUint[uint64](64)

	ParseInt8  = mkParseInt[int8](8)
	ParseInt16 = mkParseInt[int16](16)
	ParseInt32 = mkParseInt[int32](32)
	ParseInt64 = mkParseInt[int64](64)

	ParseFloat32 = mkParseFloat[float32](32)
	ParseFloat64 = mkParseFloat[float64](64)
)

func mkParseUint[T Uint](bits int) func(s string) (T, error) {
	return func(s string) (T, error) {
		return conv[uint64, T](strconv.ParseUint(s, 0, bits))
	}
}

func mkParseInt[T Int](bits int) func(s string) (T, error) {
	return func(s string) (T, error) {
		return conv[int64, T](strconv.ParseInt(s, 0, bits))
	}
}

func mkParseFloat[T Float](bits int) func(s string) (T, error) {
	return func(s string) (T, error) {
		return conv[float64, T](strconv.ParseFloat(s, bits))
	}
}

func conv[I, O Number](in I, err error) (O, error) { return O(in), err }
