package spk

import (
	_ "embed"
)

//go:embed nasxunlei-armv8.spk
var Bytes []byte

//go:embed nasxunlei-armv8.txt
var Version string
