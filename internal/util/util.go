package util

import (
	"fmt"
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

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
