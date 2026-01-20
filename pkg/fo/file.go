package fo

import (
	"bytes"
	"os"
)

// Cat 读取文件，输出文件文本内容
//   - noTrim: 不删除内容前后空白（默认是删除的）
//   - fallbackFiles: 找不到文件时，尝试读取这些文件
func Cat(file string, trim bool, fallbackFiles ...string) string {
	d, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		for _, fn := range fallbackFiles {
			if d, err = os.ReadFile(fn); !os.IsNotExist(err) {
				break
			}
		}
	}

	if err != nil {
		return ""
	}

	if trim {
		d = bytes.TrimSpace(d)
	}
	return string(d)
}
