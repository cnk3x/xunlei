package lod

// Array creates a slice from the given items.
func Array[T any](items ...T) []T {
	return items
}

// Must panics if the error is not nil.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// May creates a function that ignores the error.
func May[T any](v T, _ error) T { return v }

// CompareChain creates a compare function from the given compare functions.
func CompareChain[T any](compares ...func(a, b T) int) func(a, b T) int {
	return func(a, b T) int {
		for _, compare := range compares {
			if r := compare(a, b); r != 0 {
				return r
			}
		}
		return 0
	}
}
