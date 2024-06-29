package flags

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var FmtPrintln = func(s string) { fmt.Println("  --" + s) }
var FmtPrintf = func(s string, args ...any) { fmt.Printf(s+"\n", args...) }

type TestStruct struct {
	Str      string   `flag:"str,STR" default:"hello"`
	Strs     []string //`flag:"strs,YES,s" default:"strssss"`
	Int64p   *int64   `env:"INT64P" flag:"i" default:"234"`
	Durpsp   *[]*time.Duration
	Durps    []*time.Duration
	Durs     []time.Duration
	Int64s   []int64
	Uint64ps []*uint64
	Uintsp   *[]uint
	Time     time.Time
	Times    []time.Time
	Timep    *time.Time
	Hello    bool
	Dirs     PathList
}

func TestFlag(t *testing.T) {
	Default.SetVersion("1.0.0")
	time.Local = time.FixedZone("CST", 8*3600)

	args := []string{
		"--str", "a", "--str", "b",
		"--strs", "c", "--strs", "d",
		"--dirs", "a", "--dirs", "b", "--dirs", "c", "--dirs", "d:z",
		// "--int64p", "13",
		// "--durpsp", "1s", "--durpsp", "1h", "--durpsp", "1d",
		// "--durps", "1s", "--durps", "1h", "--durps", "1d",
		// "--int64s", "1", "--int64s", "2", "--int64s", "3",
		// "--uint64ps", "4", "--uint64ps", "5", "--uint64ps", "6",
		// "--uintsp", "7", "--uintsp", "8", "--uintsp", "9",
		// "--time", "2020-01-01",
		// "--times", "2021-01-01", "--times", "2021-01-02", "--times", "2021-01-03 11:12",
		// "--timep", "2022/01/01 11:12",
		// "--hello",
		// "-h",
	}

	var cfg TestStruct
	os.Setenv("STR", ":9981")
	os.Setenv("INT64P", "9982")
	cfg.Time = time.Now()

	cfg.Durs = []time.Duration{1 * time.Second, 1 * time.Hour}
	flag := New("test")
	flag.Struct(&cfg, nil)

	if err := flag.Parse(args); err != nil {
		t.Fatal(err)
	}

	Walk(&cfg, nil, func(field *FlagField, max int) {
		FmtPrintln(fmt.Sprintf("%-*s | %s", max, field.Name, strings.Join(rGet(field.Value.Ref), ", ")))
	})

	fmt.Println("args:")
	args, err := Args(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf(`"%s"`+"\n", strings.Join(args, `", "`))
}

const FMT = "| %-15s | %-17s | %-5s | %-5s | %-5s | %-5s | %-5s | %-5s |"

func TestReflectStruct(t *testing.T) {
	s := TestStruct{Times: []time.Time{time.Now()}, Str: "1", Int64s: []int64{}}
	sv := Ref(&s)
	st := sv.Type().Elem()
	sv = sv.Elem()

	metaPrint := metaPrinter(FmtPrintf)

	metaPrint("Main", sv, st)
	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)
		metaPrint(ft.Name, sv.Field(i), ft.Type)
	}
}

func TestValue(t *testing.T) {
	metaPrint := metaPrinter(FmtPrintf)

	var (
		conf      string = "hello"
		vDuration time.Duration
		vInt64    int64 = 3
		vInt      int
		vIntP     **int
		vInts     []int
		vIntsP    *[]int
		vIntPs    []*int
		vIntPsP   *[]*int
	)

	rSet(Ref(&vDuration), "24d1h2s", false)
	rSet(Ref(&vInt64), "100", false)
	rSet(Ref(&vInt), "101", false)
	rSet(Ref(&vIntP), "102", false)
	rSet(Ref(&vInts), "103", false)
	rSet(Ref(&vInts), "104", false)
	rSet(Ref(&vIntsP), "105", false)
	rSet(Ref(&vIntsP), "106", false)

	metaPrint("conf", Ref(&conf), nil)
	metaPrint("vDuration", Ref(&vDuration), nil)
	metaPrint("vInt64", Ref(&vInt64), nil)
	metaPrint("vInt", Ref(&vInt), nil)
	metaPrint("vIntP", Ref(&vIntP), nil)
	metaPrint("vInts", Ref(&vInts), nil)
	metaPrint("vIntsP", Ref(&vIntsP), nil)
	metaPrint("vIntPs", Ref(&vIntPs), nil)
	metaPrint("vIntPsP", Ref(&vIntPsP), nil)
}

func TestQueued(t *testing.T) {
	var s struct {
		Durations []time.Duration `flag:"vDuration" usage:"vDuration" default:"13s,16s"`
	}
	var _ = s
	// t.Log(strconv.Unquote(`"'vDuration','1234'"`))
	// var sf reflect.StructTag
	t.Log(strconv.UnquoteChar(`'vDuration'`, '\''))
}

func metaPrinter(printf func(string, ...any)) func(name string, rv reflect.Value, rt reflect.Type) {
	headers := []string{
		"name      ",
		"type             ",
		"ptr",
		"set",
		"addr",
		"valid",
		"nil",
		"zero",
	}

	headerPrint := sync.OnceFunc(func() {
		printf("| %s |", strings.Join(headers, " | "))
		printf("|%s|-------", strings.Join(sliceMap(headers, func(s string) string { return strings.Repeat("-", len(s)+2) }), "|"))
	})

	return func(name string, rv reflect.Value, rt reflect.Type) {
		headerPrint()

		if rt == nil {
			rt = rv.Type()
		}
		printf(
			"|"+strings.Repeat(" %-*s |", len(headers))+" %s",
			len(headers[0]),
			name,
			len(headers[1]),
			rt.String(),
			len(headers[2]),
			sBool(isPtr(rv)),
			len(headers[3]),
			sBool(rv.CanSet()),
			len(headers[4]),
			sBool(rv.CanAddr()),
			len(headers[5]),
			sBool(rv.IsValid()),
			len(headers[6]),
			sBool(isNil(rv)),
			len(headers[7]),
			sBool(rv.IsZero()),
			rGet(rv),
		)
	}
}

func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	default:
		return !v.IsValid()
	}
}

func isPtr(v reflect.Value) bool {
	return v.Kind() == reflect.Pointer
}

func sBool(b bool) string {
	if b {
		return " √"
	}
	return " \u00d7"
	// return "×"
}

func sliceMap[S any, R any](s []S, f func(S) R) []R {
	r := make([]R, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}
