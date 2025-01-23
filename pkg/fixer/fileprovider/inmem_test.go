package fileprovider

import (
	"path/filepath"
	"testing"

	"github.com/styrainc/regal/internal/testutil"
)

func TestFromFS(t *testing.T) {
	t.Parallel()

	tempDir := testutil.TempDirectoryOf(t, map[string]string{
		"foo/bar/baz": "bar",
		"bar/foo":     "baz",
	})
	fp := testutil.Must(NewInMemoryFileProviderFromFS([]string{
		filepath.Join(tempDir, "foo", "bar", "baz"),
		filepath.Join(tempDir, "bar", "foo"),
	}...))(t)

	if fc, err := fp.Get(filepath.Join(tempDir, "foo", "bar", "baz")); err != nil || fc != "bar" {
		t.Fatalf("expected %s, got %s", "bar", fc)
	}
}

func TestRenameConflict(t *testing.T) {
	t.Parallel()

	fp := NewInMemoryFileProvider(map[string]string{
		"/foo/bar/baz": "bar",
		"/bar/foo":     "baz",
	})

	err := fp.Rename("/foo/bar/baz", "/bar/foo")
	if err == nil {
		t.Fatal("expected error")
	}

	exp := `rename conflict: "/foo/bar/baz" cannot be renamed as the target location "/bar/foo" already exists`

	if got := err.Error(); got != exp {
		t.Fatalf("expected %s, got %s", exp, got)
	}
}
