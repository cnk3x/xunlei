package utils

import "strconv"

type uintT interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type intT interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

func StrParser[R uintT | intT, T1, T2 any, R64 uint64 | int64](f func(string, T1, T2) (R64, error), t1 T1, t2 T2) func(string) (R, error) {
	return func(s string) (R, error) {
		r, err := f(s, t1, t2)
		if err != nil {
			return 0, err
		}
		return R(r), nil
	}
}

var (
	ParseUint   = StrParser[uint](strconv.ParseUint, 0, 64)
	ParseUint8  = StrParser[uint8](strconv.ParseUint, 0, 8)
	ParseUint16 = StrParser[uint16](strconv.ParseUint, 0, 16)
	ParseUint32 = StrParser[uint32](strconv.ParseUint, 0, 32)
	ParseUint64 = StrParser[uint64](strconv.ParseUint, 0, 64)
	ParseInt    = StrParser[int](strconv.ParseInt, 0, 64)
	ParseInt8   = StrParser[int8](strconv.ParseInt, 0, 8)
	ParseInt16  = StrParser[int16](strconv.ParseInt, 0, 16)
	ParseInt32  = StrParser[int32](strconv.ParseInt, 0, 32)
	ParseInt64  = StrParser[int64](strconv.ParseInt, 0, 64)
)
