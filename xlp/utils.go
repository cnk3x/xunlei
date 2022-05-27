package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

func init() {
	if getEnv("XL_DEBUG", "0") == "1" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}
}

func fatalErr(err error, msg ...any) {
	if err != nil {
		msg = append(msg, err.Error())
		_ = log.Default().Output(2, fmt.Sprint(msg...))
		os.Exit(1)
	}
}

func nilErr(err error, msg ...any) bool {
	if err != nil {
		msg = append(msg, err.Error())
		_ = log.Default().Output(2, fmt.Sprint(msg...))
		return false
	}
	return true
}

func getEnv(key string, defaultVal ...string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	for _, val := range defaultVal {
		if val != "" {
			return val
		}
	}
	return ""
}

func randText(size int) string {
	var d = make([]byte, size)
	n, _ := rand.Read(d)
	s := hex.EncodeToString(d[:n])
	if len(s) > size {
		s = s[:size]
	}
	return s
}
