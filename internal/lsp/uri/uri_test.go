package uri

import (
	"path/filepath"
	"testing"

	"github.com/styrainc/regal/internal/lsp/clients"
)

func TestPathToURI(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path string
		want string
	}{
		"unix simple": {
			path: "/foo/bar",
			want: "file:///foo/bar",
		},
		"unix prefixed with file:// already": {
			path: "file:///foo/bar",
			want: "file:///foo/bar",
		},
		"windows not encoded": {
			path: "c:/foo/bar",
			want: "file:///c:/foo/bar",
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			t.Parallel()

			got := FromPath(clients.IdentifierGeneric, tc.path)

			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPathToURI_VSCode(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path string
		want string
	}{
		"unix simple": {
			path: "/foo/bar",
			want: "file:///foo/bar",
		},
		"unix spaces": {
			path: "/foo/bar baz",
			want: "file:///foo/bar%20baz",
		},
		"unix colon in path": {
			path: "/foo/bar:baz",
			want: "file:///foo/bar%3Abaz",
		},
		"unix prefixed": {
			path: "file:///foo/bar",
			want: "file:///foo/bar",
		},
		"windows not encoded": {
			path: "c:/foo/bar",
			want: "file:///c%3A/foo/bar",
		},
		"windows not encoded extra colon in path": {
			path: "c:/foo/bar:1",
			want: "file:///c%3A/foo/bar%3A1",
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			t.Parallel()

			got := FromPath(clients.IdentifierVSCode, tc.path)

			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestURIToPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		uri  string
		want string
	}{
		"unix unprefixed": {
			uri:  "/foo/bar",
			want: filepath.FromSlash("/foo/bar"),
		},
		"unix simple": {
			uri:  "file:///foo/bar",
			want: filepath.FromSlash("/foo/bar"),
		},
		"windows not encoded": {
			uri:  "file://c:/foo/bar",
			want: filepath.FromSlash("c:/foo/bar"),
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			t.Parallel()

			got := ToPath(clients.IdentifierGeneric, tc.uri)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestURIToPath_VSCode(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		uri  string
		want string
	}{
		"unix unprefixed": {
			uri:  "/foo/bar",
			want: filepath.FromSlash("/foo/bar"),
		},
		"unix simple": {
			uri:  "file:///foo/bar",
			want: filepath.FromSlash("/foo/bar"),
		},
		"windows encoded": {
			uri:  "file:///c%3A/foo/bar",
			want: filepath.FromSlash("c:/foo/bar"),
		},
		"windows encoded uppercase drive": {
			uri:  "file:///C%3A/foo/bar",
			want: filepath.FromSlash("C:/foo/bar"),
		},
		"unix encoded with space in path": {
			uri:  "file:///Users/foo/bar%20baz",
			want: filepath.FromSlash("/Users/foo/bar baz"),
		},
		"unix encoded with colon in path": {
			uri:  "file:///Users/foo/bar%3Abaz",
			want: filepath.FromSlash("/Users/foo/bar:baz"),
		},
		// these other examples shouldn't happen, but we should handle them
		"windows not encoded": {
			uri:  "file://c:/foo/bar",
			want: filepath.FromSlash("c:/foo/bar"),
		},
		"windows not prefixed": {
			uri:  "c:/foo/bar",
			want: filepath.FromSlash("c:/foo/bar"),
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			t.Parallel()

			got := ToPath(clients.IdentifierVSCode, tc.uri)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
