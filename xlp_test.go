package xunlei

import (
	"testing"
	"time"
)

func TestXxx(t *testing.T) {
	dt, _ := time.Parse("15:04:05.00", "05:52:48.18")

	t.Log(dt.Format(time.RFC3339Nano))
}
