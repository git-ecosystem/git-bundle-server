package utils

// Utility functions for slices not built into the standard library

func Map[T any, S any](in []T, fn func(t T) S) []S {
	out := make([]S, len(in))
	for i, t := range in {
		out[i] = fn(t)
	}
	return out
}
