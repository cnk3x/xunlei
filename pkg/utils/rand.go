package utils

import (
	"crypto/rand"
	"encoding/base32"
)

var b32 = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

// RandText 生成一个指定长度的随机字符串。
func RandText(n int) (s string) {
	var d = make([]byte, (n+4)/8*5)
	_, _ = rand.Read(d)
	if s = b32.EncodeToString(d); len(s) > n {
		s = s[:n]
	}
	return
}
