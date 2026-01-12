package fo

import (
	"os"
	"testing"
)

func TestXxx(t *testing.T) {
	wRepl := func(flag int, wFlag int) int {
		return flag&^(os.O_APPEND|os.O_TRUNC|os.O_EXCL) | wFlag
	}

	const (
		TRUNC  = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
		APPEND = os.O_CREATE | os.O_WRONLY | os.O_APPEND
		EXCL   = os.O_CREATE | os.O_WRONLY | os.O_EXCL
	)

	t.Log(wRepl(TRUNC, os.O_APPEND), APPEND)
	t.Log(wRepl(APPEND, os.O_EXCL), EXCL)
	t.Log(wRepl(EXCL, os.O_TRUNC), TRUNC)
}
