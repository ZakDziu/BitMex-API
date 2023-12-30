package utils

// Mod applies given modification to any value.
//
//nolint:ireturn,nolintlint // By design
func Mod[T any](v T, f func(*T)) T {
	f(&v)

	return v
}
