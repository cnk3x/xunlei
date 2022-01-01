package main

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/fatih/color"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

func Standard(prefix string, writers ...io.Writer) Logger {
	var w io.Writer
	if len(writers) > 1 {
		w = io.MultiWriter(writers...)
	} else if len(writers) == 1 {
		w = writers[0]
	}
	return &std{prefix: prefix, out: w}
}

type std struct {
	prefix string
	out    io.Writer
	once   sync.Once
}

func (l *std) output(tag, format string, args ...interface{}) {
	l.once.Do(func() {
		if l.out == nil {
			l.out = os.Stderr
		}
	})
	if format == l.prefix {
		format = ""
	}

	sPrinter := fmt.Sprintf
	switch tag {
	case "DEBU":
	case "INFO":
		sPrinter = color.GreenString
	case "WARN":
		sPrinter = color.BlueString
	case "FATA":
		sPrinter = color.RedString
	}

	// prefix := sPrinter("[%s] ", tag) + time.Now().Format(time.RFC3339) + sPrinter(" [%s] ", l.prefix)
	fmt.Fprintf(l.out, sPrinter("[%s] ", l.prefix)+format+"\n", args...)
}

func (l *std) Debugf(format string, args ...interface{}) {
	l.output("DEBU", format, args...)
}

func (l *std) Infof(format string, args ...interface{}) {
	l.output("INFO", format, args...)
}

func (l *std) Warnf(format string, args ...interface{}) {
	l.output("WARN", format, args...)
}

func (l *std) Fatalf(format string, args ...interface{}) {
	l.output("FATA", format, args...)
	os.Exit(1)
}
