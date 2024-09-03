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
		err := os.MkdirAll(filepath.Dir(file), 0o755)
		if err != nil {
			t.Fatal(err)
		}

		// write the file
		err = os.WriteFile(file, []byte(content), 0o600)
		if err != nil {
			t.Fatal(err)
		}
	}

	fp, err := NewInMemoryFileProviderFromFS(util.Keys(files)...)
	if err != nil {
		t.Fatal(err)
	}

	if fc, err := fp.Get(tempDir + "/foo/bar/baz"); err != nil || string(fc) != "bar" {
		t.Fatalf("expected %s, got %s", "bar", string(fc))
	}
}

func TestRenameConflict(t *testing.T) {
	t.Parallel()

	fileA := "/foo/bar/baz"
	fileB := "/bar/foo"

	files := map[string][]byte{
		fileA: []byte("bar"),
		fileB: []byte("baz"),
	}

	fp := NewInMemoryFileProvider(files)

	err := fp.Rename(fileA, fileB)
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "rename conflict: file /bar/foo already exists" {
		t.Fatalf("expected %s, got %s", "rename conflict", err.Error())
	}
}
