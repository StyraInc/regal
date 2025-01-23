package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func Must[T any](x T, err error) func(t *testing.T) T {
	return func(t *testing.T) T {
		t.Helper()

		if err != nil {
			t.Fatal(err)
		}

		return x
	}
}

func TempDirectoryOf(t *testing.T, files map[string]string) string {
	t.Helper()

	tmpDir := t.TempDir()

	for file, contents := range files {
		path := filepath.Join(tmpDir, file)

		MustMkdirAll(t, filepath.Dir(path))
		MustWriteFile(t, path, []byte(contents))
	}

	return tmpDir
}

func MustMkdirAll(t *testing.T, path ...string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(path...), 0o755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

func MustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}

	return contents
}

func MustWriteFile(t *testing.T, path string, contents []byte) {
	t.Helper()

	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func MustRemove(t *testing.T, path string) {
	t.Helper()

	if err := os.Remove(path); err != nil {
		t.Fatalf("failed to remove file %s: %v", path, err)
	}
}
