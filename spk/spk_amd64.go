package spk

import (
	_ "embed"
)

//go:embed nasxunlei-x86_64.spk
var Bytes []byte

//go:embed nasxunlei-x86_64.txt
var Version string
