package util

// Filter returns a new slice containing only the elements of s that
// satisfy the predicate f. This function runs each element of s through
// f twice in order to allocate exactly what is needed. This is commonly
// *much* more efficient than appending blindly, but do not use this if
// the predicate function is expensive to compute.
func Filter[T any](s []T, f func(T) bool) []T {
	n := 0

	for i := range s {
		if f(s[i]) {
			n++
		}
	}

	if n == 0 {
		return nil
	}

	r := make([]T, 0, n)

	for i := range s {
		if f(s[i]) {
			r = append(r, s[i])
		}
	}

	return r
}

// Map applies the function f to each element of s and returns a new slice.
func Map[T, R any](s []T, f func(T) R) []R {
	r := make([]R, len(s))

	for i := range s {
		r[i] = f(s[i])
	}

	return r
}
