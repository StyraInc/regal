package util

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestFindClosestMatchingRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		roots    []string
		path     string
		expected string
	}{
		{
			name:     "different length roots",
			roots:    []string{"/a/b/c", "/a/b", "/a"},
			path:     "/a/b/c/d/e/f",
			expected: "/a/b/c",
		},
		{
			name:     "exact match",
			roots:    []string{"/a/b/c", "/a/b", "/a"},
			path:     "/a/b",
			expected: "/a/b",
		},
		{
			name:     "mixed roots",
			roots:    []string{"/a/b/c/b/a", "/c/b", "/a/d/c/f"},
			path:     "/c/b/a",
			expected: "/c/b",
		},
		{
			name:     "no matching root",
			roots:    []string{"/a/b/c"},
			path:     "/d",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := FindClosestMatchingRoot(test.path, test.roots)
			if got != test.expected {
				t.Fatalf("expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestDirCleanUpPaths(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		State                     map[string]string
		DeleteTarget              string
		AdditionalPreserveTargets []string
		Expected                  []string
	}{
		"simple": {
			DeleteTarget: "foo/bar.rego",
			State: map[string]string{
				"foo/bar.rego": "package foo",
			},
			Expected: []string{"foo"},
		},
		"not empty": {
			DeleteTarget: "foo/bar.rego",
			State: map[string]string{
				"foo/bar.rego": "package foo",
				"foo/baz.rego": "package foo",
			},
			Expected: []string{},
		},
		"all the way up": {
			DeleteTarget: "foo/bar/baz/bax.rego",
			State: map[string]string{
				"foo/bar/baz/bax.rego": "package baz",
			},
			Expected: []string{"foo/bar/baz", "foo/bar", "foo"},
		},
		"almost all the way up": {
			DeleteTarget: "foo/bar/baz/bax.rego",
			State: map[string]string{
				"foo/bar/baz/bax.rego": "package baz",
				"foo/bax.rego":         "package foo",
			},
			Expected: []string{"foo/bar/baz", "foo/bar"},
		},
		"with preserve targets": {
			DeleteTarget: "foo/bar/baz/bax.rego",
			AdditionalPreserveTargets: []string{
				"foo/bar/baz_test/bax.rego",
			},
			State: map[string]string{
				"foo/bar/baz/bax.rego": "package baz",
				"foo/bax.rego":         "package foo",
			},
			// foo/bar is not deleted because of the preserve target
			Expected: []string{"foo/bar/baz"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()

			for k, v := range test.State {
				if err := os.MkdirAll(filepath.Dir(filepath.Join(tempDir, k)), 0o755); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if err := os.WriteFile(filepath.Join(tempDir, k), []byte(v), 0o600); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			expected := make([]string, len(test.Expected))
			for i, v := range test.Expected {
				expected[i] = filepath.Join(tempDir, v)
			}

			additionalPreserveTargets := []string{tempDir}
			for i, v := range test.AdditionalPreserveTargets {
				additionalPreserveTargets[i] = filepath.Join(tempDir, v)
			}

			got, err := DirCleanUpPaths(filepath.Join(tempDir, test.DeleteTarget), additionalPreserveTargets)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !slices.Equal(got, expected) {
				t.Fatalf("expected\n%v\ngot:\n%v", strings.Join(expected, "\n"), strings.Join(got, "\n"))
			}
		})
	}
}
