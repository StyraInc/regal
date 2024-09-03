package util

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Keys returns slice of keys from map.
func Keys[K comparable, V any](m map[K]V) []K {
	ks := make([]K, len(m))
	i := 0

	for k := range m {
		ks[i] = k
		i++
	}

	return ks
}

// Contains checks if slice contains element.
func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}

	return false
}

// NullToEmpty returns empty slice if provided slice is nil.
func NullToEmpty[T any](a []T) []T {
	if a == nil {
		return []T{}
	}

	return a
}

// SearchMap searches map for value at provided path.
func SearchMap(object map[string]any, path []string) (any, error) {
	current := object
	traversed := make([]string, 0, len(path))

	for i, p := range path {
		var ok bool
		if i == len(path)-1 {
			value, ok := current[p]
			if ok {
				return value, nil
			}

			return nil, fmt.Errorf("no '%v' attribute at path '%v'", p, strings.Join(traversed, "."))
		}

		if current, ok = current[p].(map[string]any); !ok {
			return nil, fmt.Errorf("no '%v' attribute at path '%v'", p, strings.Join(traversed, "."))
		}

		traversed = append(traversed, p)
	}

	return current, nil
}

// Must takes a value and an error (as commonly returned by Go functions) and panics if the error is not nil.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

// Map applies a function to each element of a slice and returns a new slice with the results.
func Map[T, U any](f func(T) U, a []T) []U {
	b := make([]U, len(a))

	for i, v := range a {
		b[i] = f(v)
	}

	return b
}

// MapInvert inverts a map (swaps keys and values).
func MapInvert[K comparable, V comparable](m map[K]V) map[V]K {
	n := make(map[V]K, len(m))
	for k, v := range m {
		n[v] = k
	}

	return n
}

// FindClosestMatchingRoot finds the closest matching root for a given path.
// If no matching root is found, an empty string is returned.
func FindClosestMatchingRoot(path string, roots []string) string {
	currentLongestSuffix := 0
	longestSuffixIndex := -1

	for i, root := range roots {
		if root == path {
			return path
		}

		if !strings.HasPrefix(path, root) {
			continue
		}

		suffix := strings.TrimPrefix(root, path)

		if len(suffix) > currentLongestSuffix {
			currentLongestSuffix = len(suffix)
			longestSuffixIndex = i
		}
	}

	if longestSuffixIndex == -1 {
		return ""
	}

	return roots[longestSuffixIndex]
}

// FilepathJoiner returns a function that joins provided path with base path.
func FilepathJoiner(base string) func(string) string {
	return func(path string) string {
		return filepath.Join(base, path)
	}
}
