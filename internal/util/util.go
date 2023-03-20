package util

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
