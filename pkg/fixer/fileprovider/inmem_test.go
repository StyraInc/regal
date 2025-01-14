package fileprovider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/styrainc/regal/internal/util"
)

func TestFromFS(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	files := map[string]string{
		tempDir + "/foo/bar/baz": "bar",
		tempDir + "/bar/foo":     "baz",
	}

	for file, content := range files {
		/// make the dir
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}

		// write the file
		if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	fp, err := NewInMemoryFileProviderFromFS(util.Keys(files)...)
	if err != nil {
		t.Fatal(err)
	}

	if fc, err := fp.Get(tempDir + "/foo/bar/baz"); err != nil || fc != "bar" {
		t.Fatalf("expected %s, got %s", "bar", fc)
	}
}

func TestRenameConflict(t *testing.T) {
	t.Parallel()

	fileA := "/foo/bar/baz"
	fileB := "/bar/foo"

	files := map[string]string{
		fileA: "bar",
		fileB: "baz",
	}

	fp := NewInMemoryFileProvider(files)

	err := fp.Rename(fileA, fileB)
	if err == nil {
		t.Fatal("expected error")
	}

	exp := `rename conflict: "/foo/bar/baz" cannot be renamed as the target location "/bar/foo" already exists`

	if got := err.Error(); got != exp {
		t.Fatalf("expected %s, got %s", exp, got)
	}
}
