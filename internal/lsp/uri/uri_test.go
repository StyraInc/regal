package uri

import (
	"testing"

	"github.com/styrainc/regal/internal/lsp/clients"
)

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
		"unix prefixed": {
			path: "file:///foo/bar",
			want: "file:///foo/bar",
		},
		"windows encoded": {
			path: "c%3A/foo/bar",
			want: "file:///c%3A/foo/bar",
		},
		"windows not encoded": {
			path: "c:/foo/bar",
			want: "file:///c%3A/foo/bar",
		},
	}

	for label, tc := range testCases {
		tt := tc

		t.Run(label, func(t *testing.T) {
			t.Parallel()
			got := FromPath(clients.IdentifierVSCode, tt.path)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
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
			want: "/foo/bar",
		},
		"unix simple": {
			uri:  "file:///foo/bar",
			want: "/foo/bar",
		},
		"windows encoded": {
			uri:  "file:///c%3A/foo/bar",
			want: "c:/foo/bar",
		},
		// these other examples shouldn't happen, but we should handle them
		"windows not encoded": {
			uri:  "file://c:/foo/bar",
			want: "c:/foo/bar",
		},
		"windows not prefixed": {
			uri:  "c:/foo/bar",
			want: "c:/foo/bar",
		},
		"windows not prefixed, but encoded": {
			uri:  "c%3A/foo/bar",
			want: "c:/foo/bar",
		},
	}

	for label, tc := range testCases {
		tt := tc

		t.Run(label, func(t *testing.T) {
			t.Parallel()

			got := ToPath(clients.IdentifierVSCode, tt.uri)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
