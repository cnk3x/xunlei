package flags

import (
	"reflect"
	"strings"
)

type Value struct {
	Ref     reflect.Value //引用对象
	typ     reflect.Type  //引用类型
	defs    []string      //默认值字符串
	changed bool
}

func newValue(v reflect.Value) *Value { return &Value{Ref: v, typ: v.Type(), defs: rGet(v)} }

func (v *Value) Type() string { return rType(v.DirectType()) }

func (v *Value) Set(s string) (err error) {
	if err = rSet(v.Ref, s, !v.changed); err != nil {
		return
	}
	v.changed = true
	return
}

func (v *Value) String() string {
	if len(v.defs) > 0 {
		if v.IsKind(reflect.Slice) {
			return "[" + strings.Join(v.defs, ",") + "]"
		} else {
			return v.defs[0]
		}
	}
	return ""
}

func (v *Value) SetString(args []string, asDefault bool) (err error) {
	for i, arg := range args {
		if err = rSet(v.Ref, arg, i == 0); err != nil {
			return
		}
	}

	if asDefault {
		v.defs = args
	}

	v.changed = false

	return
}

func (v *Value) DirectType() reflect.Type      { return typeIndirect(v.typ) }
func (v *Value) IsKind(kind reflect.Kind) bool { return v.DirectType().Kind() == kind }
