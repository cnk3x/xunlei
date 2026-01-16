package sys

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cnk3x/xunlei/pkg/vms/sys/errs"
)

func TestMkdir(t *testing.T) {
	var err error
	os.MkdirAll("test_data/EEXIST", 0777)

	err = os.Mkdir("test_data/EEXIST", 0777) //存在 EEXIST
	printErr(t, err)
	err = os.Mkdir("test_data/FEXIST", 0777) //路径存在, 且非文件夹, EEXIST
	printErr(t, err)
	err = os.Mkdir("test_data/ENOENT/1", 0777) //上级路径不存在 ENOENT
	printErr(t, err)
	err = os.Mkdir("test_data/FEXIST/1", 0777) //上级路径存在, 且非文件夹, ENOTDIR
	printErr(t, err)
	err = os.Mkdir("test_data/n1", 0777) //路径存在, 为软链接，软链接目标丢失 EEXIST
	printErr(t, err)
	err = os.Mkdir("test_data/n2", 0777) //路径存在, 为软链接，软链接目标存在 EEXIST
	printErr(t, err)
	err = os.Mkdir("test_data/n3", 0777) //路径存在, 为软链接，软链接目标存在,且非文件夹 EEXIST
	printErr(t, err)

	l1, _ := os.Readlink("test_data/n1")
	l2, _ := os.Readlink("test_data/n2")
	l3, _ := os.Readlink("test_data/n3")
	l4, _ := os.Readlink("test_data/n4")

	t1, _ := filepath.EvalSymlinks("test_data/n1")
	t2, _ := filepath.EvalSymlinks("test_data/n2")
	t3, _ := filepath.EvalSymlinks("test_data/n3")
	t4, _ := filepath.EvalSymlinks("test_data/n4")

	t.Logf("l: n1: %s, n2: %s, n3: %s, n4: %s", l1, l2, l3, l4)
	t.Logf("t: n1: %s, n2: %s, n3: %s, n4: %s", t1, t2, t3, t4)
}

// switch {
// case errors.Is(err, syscall.ENOENT): //上级路径不存在
// case errors.Is(err, syscall.ENOTDIR): //上级路径存在, 且非文件夹
// case errors.Is(err, syscall.EEXIST): //路径存在, 可能是文件夹，可能是文件
// }
func printErr(t *testing.T, err error) {
	if name, code, ok := errs.ErrCode(err); ok {
		t.Logf("[%T(%s(%s))]%v", err, name, code, err)
	} else {
		t.Logf("[%T]%v", err, err)
	}
}
