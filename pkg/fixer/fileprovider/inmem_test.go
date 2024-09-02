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

	if fc, err := fp.GetFile(tempDir + "/foo/bar/baz"); err != nil || string(fc) != "bar" {
		t.Fatalf("expected %s, got %s", "bar", string(fc))
	}
}
