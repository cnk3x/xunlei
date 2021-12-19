package main

import (
	"bytes"
	"os"
	"path/filepath"
)

func getDownloadDIR() string {
	downloadDIR, _ := os.ReadFile(filepath.Join(SYNOPKG_PKGBASE, ".downloadDIR"))
	downloadDIR = bytes.TrimSpace(downloadDIR)
	return string(downloadDIR)
}

func saveDownloadDIR() string {
	downloadDIR, _ := os.ReadFile(filepath.Join(SYNOPKG_PKGBASE, ".downloadDIR"))
	downloadDIR = bytes.TrimSpace(downloadDIR)
	return string(downloadDIR)
}
