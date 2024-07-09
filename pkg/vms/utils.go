package vms

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// ResolvePath 解析给定的基路径和一系列路径列表，返回它们的绝对路径、相对路径、非子路径和错误信息。
//
//	base 是基础路径。
//	list 是待解析的路径列表。
//	allPaths 是所有路径的绝对路径数组 outPaths + subPaths。
//	outPaths 是不在基础路径下的路径绝对路径数组。
//	subPaths 是在基础路径下的路径去掉基础路径前缀的绝对路径数组。
//	err 是错误信息。
func ResolvePath(base string, list ...string) (allPaths, outPaths, subPaths []string, err error) {
	// 如果基础路径为空，则设置为根路径。
	if base == "" {
		base = "/"
	} else if base, err = filepath.Abs(base); err != nil {
		// 如果获取基础路径的绝对路径失败，则返回错误。
		return
	}

	// 如果路径列表为空，则返回错误。
	if len(list) == 0 {
		err = fmt.Errorf("no path to resolve")
		return
	}

	// 遍历路径列表。
	for _, p := range list {
		// 如果路径为空，则返回错误。
		if p == "" {
			err = fmt.Errorf("path is empty")
			return
		}

		// 获取路径的绝对路径。
		if p, err = filepath.Abs(p); err != nil {
			// 如果获取路径的绝对路径失败，则返回错误。
			return
		}

		// 检查是否存在重复的绝对路径。
		if slices.Contains(allPaths, p) {
			err = fmt.Errorf("duplicate path: %s", p)
			return
		}

		// 如果基础路径是根路径，则直接添加绝对路径到结果中。
		if base == "/" {
			allPaths = append(allPaths, p)
			subPaths = append(subPaths, p)
			continue
		}

		// 获取路径相对于基础路径的相对路径。
		var rel string
		if rel, err = filepath.Rel(base, p); err != nil {
			// 如果获取相对路径失败，则返回错误。
			return
		}

		// 判断路径是否是基础路径的子路径。
		isSub := !strings.Contains(rel, "..")
		// 根据是否是子路径，构造相对路径。
		if isSub {
			p = "/" + rel
		}

		// 检查是否存在重复的相对路径。
		if slices.Contains(allPaths, p) {
			err = fmt.Errorf("duplicate path: %s", p)
			return
		}

		// 根据路径是否是基础路径的子路径，分类处理。
		// 如果是子路径，添加到 subPaths 数组；如果不是，添加到 outPaths 数组。
		if isSub {
			subPaths = append(subPaths, p)
		} else {
			outPaths = append(outPaths, p)
		}
		// 将相对路径添加到 allPaths 数组。
		allPaths = append(allPaths, p)
	}
	// 返回处理结果。
	return
}
