package cmd

import (
	"runtime"
	"slices"
	"strings"
)

/* environments functions */

// EnvSet 环境变量集合，封装了一些对环境变量集合的操作
type EnvSet []string

// Has 判断是否已经包含某个键
func (src EnvSet) Has(k string) bool { return slices.ContainsFunc(src, findPrefix(k+"=")) }

// Set 设置环境变量，如果已经存在则先删除再追加
func (src EnvSet) Set(k, v string) (out EnvSet) {
	out = src
	if k = strings.TrimSpace(k); k != "" {
		out = append(out.Unset(k), k+"="+v)
	}
	return out
}

// Unset 删除环境变量
func (src EnvSet) Unset(k string) EnvSet { return slices.DeleteFunc(src, findPrefix(k+"=")) }

// Clean 清理无效的环境变量，重复的保留最后一个
func (src EnvSet) Clean() EnvSet {
	var j = len(src)
	for i := len(src) - 1; i >= 0; i-- {
		if kv := strings.SplitN(src[i], "=", 2); len(kv) == 2 && kv[0] != "" {
			if !src[j:].Has(kv[0]) {
				j--
				src[j] = src[i]
			}
		}
	}
	clear(src[:j])
	return src[j:]
}

// 查找前缀，如果是windows环境，忽略大小写
func findPrefix(find string) func(src string) bool {
	return func(src string) bool {
		if runtime.GOOS == "windows" {
			return len(src) >= len(find) && strings.EqualFold(src[:len(find)], find)
		}
		return len(src) >= len(find) && src[:len(find)] == find
	}
}

/* environments functions end*/
