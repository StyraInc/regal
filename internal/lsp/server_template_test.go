package lsp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/uri"
)

func TestTemplateContentsForFile(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		FileKey           string
		CacheFileContents string
		DiskContents      map[string]string
		RequireConfig     bool
		ExpectedContents  string
		ExpectedError     string
	}{
		"existing contents in file": {
			FileKey:           "foo/bar.rego",
			CacheFileContents: "package foo",
			ExpectedError:     "file already has contents",
		},
		"existing contents on disk": {
			FileKey:           "foo/bar.rego",
			CacheFileContents: "",
			DiskContents: map[string]string{
				"foo/bar.rego": "package foo",
			},
			ExpectedError: "file on disk already has contents",
		},
		"empty file is templated as main when no root": {
			FileKey:           "foo/bar.rego",
			CacheFileContents: "",
			DiskContents: map[string]string{
				"foo/bar.rego": "",
			},
			ExpectedContents: "package main\n\nimport rego.v1\n",
		},
		"empty file is templated based on root": {
			FileKey:           "foo/bar.rego",
			CacheFileContents: "",
			DiskContents: map[string]string{
				"foo/bar.rego":       "",
				".regal/config.yaml": "",
			},
			ExpectedContents: "package foo\n\nimport rego.v1\n",
		},
		"empty test file is templated based on root": {
			FileKey:           "foo/bar_test.rego",
			CacheFileContents: "",
			DiskContents: map[string]string{
				"foo/bar_test.rego":  "",
				".regal/config.yaml": "",
			},
			RequireConfig:    true,
			ExpectedContents: "package foo_test\n\nimport rego.v1\n",
		},
		"empty deeply nested file is templated based on root": {
			FileKey:           "foo/bar/baz/bax.rego",
			CacheFileContents: "",
			DiskContents: map[string]string{
				"foo/bar/baz/bax.rego": "",
				".regal/config.yaml":   "",
			},
			ExpectedContents: "package foo.bar.baz\n\nimport rego.v1\n",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			td := t.TempDir()

			// init the state on disk
			for f, c := range tc.DiskContents {
				dir := filepath.Dir(f)

				err := os.MkdirAll(filepath.Join(td, dir), 0o755)
				if err != nil {
					t.Fatalf("failed to create directory %s: %s", dir, err)
				}

				err = os.WriteFile(filepath.Join(td, f), []byte(c), 0o600)
				if err != nil {
					t.Fatalf("failed to write file %s: %s", f, err)
				}
			}

			// create a new language server
			s := NewLanguageServer(&LanguageServerOptions{ErrorLog: newTestLogger(t)})
			s.workspaceRootURI = uri.FromPath(clients.IdentifierGeneric, td)

			fileURI := uri.FromPath(clients.IdentifierGeneric, filepath.Join(td, tc.FileKey))

			s.cache.SetFileContents(fileURI, tc.CacheFileContents)

			newContents, err := s.templateContentsForFile(fileURI)
			if tc.ExpectedError != "" {
				if !strings.Contains(err.Error(), tc.ExpectedError) {
					t.Fatalf("expected error to contain %q, got %q", tc.ExpectedError, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if newContents != tc.ExpectedContents {
				t.Fatalf("expected contents to be\n%s\ngot\n%s", tc.ExpectedContents, newContents)
			}
		})
	}
}
