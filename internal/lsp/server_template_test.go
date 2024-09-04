package lsp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/uri"
)

func TestServerTemplateContentsForFile(t *testing.T) {
	t.Parallel()

	s := NewLanguageServer(
		&LanguageServerOptions{
			ErrorLog: os.Stderr,
		},
	)

	td := t.TempDir()

	filePath := filepath.Join(td, "foo/bar/baz.rego")
	regalPath := filepath.Join(td, ".regal/config.yaml")

	initialState := map[string]string{
		filePath:  "",
		regalPath: "",
	}

	// create the initial state needed for the regal config root detection
	for file := range initialState {
		fileDir := filepath.Dir(file)

		err := os.MkdirAll(fileDir, 0o755)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = os.WriteFile(file, []byte(""), 0o600)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	fileURI := uri.FromPath(clients.IdentifierGeneric, filePath)

	s.cache.SetFileContents(fileURI, "")

	newContents, err := s.templateContentsForFile(fileURI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if newContents != "package foo.bar\n\nimport rego.v1\n" {
		t.Fatalf("unexpected contents: %v", newContents)
	}
}
