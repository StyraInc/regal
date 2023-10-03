package testutil

import "testing"

func Must[T any](x T, err error) func(t *testing.T) T {
	return func(t *testing.T) T {
		t.Helper()

		if err != nil {
			t.Fatal(err)
		}

		return x
	}
}
