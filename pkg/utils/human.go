package utils

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
)

func HumanBytes[T UintT | IntT](n T, prec ...int) string {
	if f := float64(n); f >= 1024 {
		for i, u := range slices.Backward([]rune("KMGTE")) {
			if base := float64(int64(1) << (10 * (i + 1))); float64(n) >= base {
				return fmt.Sprintf("%s %ciB", strconv.FormatFloat(f/base, 'f', cmp.Or(cmp.Or(prec...), 2), 64), u)
			}
		}
	}
	return fmt.Sprintf("%d bytes", n)
}
