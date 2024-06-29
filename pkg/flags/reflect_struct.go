package flags

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"
)

type FlagField struct {
	Field           reflect.StructField
	Name            string
	Shorthand       string
	Usage           string
	Value           *Value
	Env             []string
	Deprecated      string
	ShortDeprecated string

	defTag string
}

func (f *FlagField) applyPrefix(prefix *childPrefix) *FlagField {
	if prefix != nil {
		f.Name = prefix.Flag + f.Name
		for i, env := range f.Env {
			f.Env[i] = prefix.Env + env
		}
	}
	return f
}

func (f *FlagField) applyDefault() (err error) {
	if f.defTag != "" {
		err = f.Value.SetString(o2s(f.defTag), true)
	}

	for _, k := range f.Env {
		if s := os.Getenv(strings.TrimSpace(k)); s != "" {
			if f.defTag != "" {
				f.Value.defs = o2s(f.defTag)
			}
			err = f.Value.SetString(o2s(s), false)
			return
		}
	}
	return
}

func toFields(src any, prefix *childPrefix) (fields []*FlagField, err error) {
	r := Ref(src)

	if !r.CanSet() {
		err = fmt.Errorf("can't set %T", src)
		return
	}

	prefix = sels(prefix, prefixDefault)

	for i, t := 0, r.Type(); i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		switch {
		case isAllow(f.Type):
			item, e := parseField(r, f, i)
			if e != nil {
				if e == errSkip {
					continue
				}
				err = e
				return
			}
			fields = append(fields, item.applyPrefix(prefix))
		case isStructType(f.Type):
			var cFields []*FlagField
			if f.Anonymous {
				cFields, err = toFields(mkPtr(r.Field(i)), prefix)
			} else {
				np, e := parseChildPrefix(f)
				if e != nil {
					if e == errSkip {
						continue
					}
					err = e
					return
				}

				np.Flag, np.Env = prefix.Flag+np.Flag, prefix.Env+np.Env
				cFields, err = toFields(mkPtr(r.Field(i)), np)
			}

			if err != nil {
				return
			}

			fields = append(fields, cFields...)
		}
	}

	return
}

var errSkip = errors.New("skip parse this field")

func parseField(r reflect.Value, f reflect.StructField, fieldIndex int) (item FlagField, err error) {
	if item.Name, item.Shorthand, item.Env, err = parseTag(f, false); err != nil {
		return
	}

	item.Field = f
	item.Value = newValue(r.Field(fieldIndex))
	item.Usage = getTag(f.Tag, _TAG_USAGE)
	item.defTag = getTag(f.Tag, _TAG_DEFAULT)

	if deprecatedTag := getTag(f.Tag, _TAG_DEPRECATED); deprecatedTag != "" {
		nn := fieldSpilt(deprecatedTag)
		for _, n := range nn {
			if n != "" {
				if item.Deprecated == "" {
					item.Deprecated = n
				} else if item.ShortDeprecated == "" {
					item.ShortDeprecated = n
				}
			}
		}
		if len(item.Deprecated) <= 1 && len(item.ShortDeprecated) > 1 {
			item.Deprecated, item.ShortDeprecated = item.ShortDeprecated, item.Deprecated
		}
	}

	return
}

func parseChildPrefix(f reflect.StructField) (*childPrefix, error) {
	name, _, envs, err := parseTag(f, true)
	if err != nil {
		return nil, err
	}
	if len(envs) == 0 {
		envs = append(envs, strings.ToUpper(name))
	}
	return &childPrefix{name + ".", envs[0] + "_"}, nil
}

func parseTag(f reflect.StructField, ignoreError bool) (name, shorthand string, envKeys []string, err error) {
	if shorthand = getTag(f.Tag, _TAG_SHORTHAND); len(shorthand) > 0 {
		if ignoreError {
			shorthand = ""
		} else {
			err = fmt.Errorf("shorthand must be a single character, got: %s", shorthand)
			return
		}
	}

	if envTag := getTag(f.Tag, _TAG_ENV); envTag != "" && envTag != "-" {
		envKeys = append(envKeys, fieldSpilt(envTag)...)
	}

	if flagTag := getTag(f.Tag, _TAG_FLAG); flagTag != "" {
		if flagTag == "-" {
			err = errSkip
			return
		}

		for _, s := range fieldSpilt(flagTag) {
			switch l := len(s); l {
			case 0:
				continue
			case 1:
				if shorthand != "" {
					if ignoreError {
						continue
					}
					err = fmt.Errorf("can only define one shorthand flag, got: %s, already: %s", s, shorthand)
					return
				}
				shorthand = s
			default:
				if name == "" {
					name = s
				} else {
					envKeys = append(envKeys, s)
				}
			}
		}
	}

	if name == "" {
		name = strings.ToLower(f.Name)
	}

	return
}

const (
	_TAG_FLAG       = "flag"
	_TAG_SHORTHAND  = "shorthand"
	_TAG_DEPRECATED = "deprecated"
	_TAG_ENV        = "env"
	_TAG_USAGE      = "usage"
	_TAG_DEFAULT    = "default"
)

var (
	fieldSpilt = func(s string) []string {
		fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' || r == '|' || unicode.IsSpace(r) })
		var x int
		for _, s := range fields {
			if s = strings.Trim(s, "-_"); s != "" {
				fields[x] = s
				x++
			}
		}
		return fields[:x]
	}

	getTag = func(tag reflect.StructTag, tagName string) string { return strings.TrimSpace(tag.Get(tagName)) }
)
