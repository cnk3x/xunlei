package vms

import (
	"testing"
)

func TestRel(t *testing.T) {
	var base = "/xunlei"

	var paths = []string{
		"/xunlei/data",
		"/xunlei/downloads",
		"/downloads",
		"/bin", "/dev",
		"/lib", "/lib64",
		"/tmp", "/run", "/proc",
		"/etc/hosts", "/etc/hostname",
		"/etc/timezone", "/etc/localtime",
		"/etc/resolv.conf", "/etc/profile",
	}

	allPaths, outPaths, subPaths, err := ResolvePath(base, paths...)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("allPaths: %v", allPaths)
	t.Logf("outPaths: %v", outPaths)
	t.Logf("subPaths %v", subPaths)
}
