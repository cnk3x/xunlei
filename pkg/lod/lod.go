package lod

// Array creates a slice from the given items.
func Array[T any](items ...T) []T {
	return items
}
