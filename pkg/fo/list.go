package fo

import (
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"strings"
)

func WalkDir(root string, fn func(path string, d fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return fn(path, d)
	})
}

func FileSeq(root string, predicates ...func(name string, fileEntry fs.DirEntry) bool) iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		OpenRead(root, Files(func(fileEntry fs.DirEntry) error {
			if !yield(fileEntry.Name(), fileEntry) {
				return fs.SkipAll
			}
			return nil
		}, predicates...))
	}
}

func Files(onFile func(fileEntry fs.DirEntry) error, predicates ...func(name string, fileEntry fs.DirEntry) bool) Process {
	return func(f *os.File) (err error) {
		var entries []fs.DirEntry
		if entries, err = f.ReadDir(-1); err != nil {
			return
		}

	LOOP_ENTRY:
		for _, entry := range entries {
			name := entry.Name()
			for _, predicate := range predicates {
				if predicate != nil && !predicate(name, entry) {
					continue LOOP_ENTRY
				}
			}
			if err = onFile(entry); err != nil {
				break
			}
		}

		if err == fs.SkipAll {
			err = nil
		}

		return
	}
}

func MatchExt(exts ...string) func(string, fs.DirEntry) bool {
	return func(name string, _ fs.DirEntry) bool { return HasExt(name, exts...) }
}

func MatchDir(exts ...string) func(string, fs.DirEntry) bool {
	return func(name string, entry fs.DirEntry) bool { return entry.IsDir() }
}

func MatchNoHidden(exts ...string) func(string, fs.DirEntry) bool {
	return func(name string, _ fs.DirEntry) bool { return strings.HasPrefix(name, ".") }
}

func HasExt(name string, exts ...string) bool {
	for _, ext := range exts {
		if len(name) >= len(ext) && strings.EqualFold(name[len(name)-len(ext):], ext) {
			return true
		}
	}
	return false
}

func StartWith(name string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if len(name) >= len(prefix) && strings.EqualFold(name[:len(prefix)], prefix) {
			return true
		}
	}
	return false
}
