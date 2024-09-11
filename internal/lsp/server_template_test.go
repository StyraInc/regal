package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/types"
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

func TestNewFileTemplating(t *testing.T) {
	t.Parallel()

	var err error

	// set up the workspace content with some example rego and regal config
	tempDir := t.TempDir()

	files := map[string]string{
		".regal/config.yaml": `rules:
  idiomatic:
    directory-package-mismatch:
      level: error
      exclude-test-suffix: false
`,
	}

	for f, fc := range files {
		err := os.MkdirAll(filepath.Dir(filepath.Join(tempDir, f)), 0o755)
		if err != nil {
			t.Fatalf("failed to create directory %s: %s", filepath.Dir(filepath.Join(tempDir, f)), err)
		}

		err = os.WriteFile(filepath.Join(tempDir, f), []byte(fc), 0o600)
		if err != nil {
			t.Fatalf("failed to write file %s: %s", f, err)
		}
	}

	// set up the server and client connections
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ls := NewLanguageServer(&LanguageServerOptions{
		ErrorLog: newTestLogger(t),
	})

	go ls.StartConfigWorker(ctx)
	go ls.StartTemplateWorker(ctx)

	receivedMessages := make(chan []byte, 10)
	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		bs, err := json.MarshalIndent(req.Params, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal params: %s", err)
		}

		receivedMessages <- bs

		return struct{}{}, nil
	}

	connServer, connClient, cleanup := createConnections(ctx, ls.Handle, clientHandler)
	defer cleanup()

	ls.SetConn(connServer)

	// 1. Client sends initialize request
	request := types.InitializeParams{
		RootURI:    fileURIScheme + tempDir,
		ClientInfo: types.Client{Name: "go test"},
	}

	var response types.InitializeResult

	err = connClient.Call(ctx, "initialize", request, &response)
	if err != nil {
		t.Fatalf("failed to send initialize request: %s", err)
	}

	// 2. Client sends initialized notification no response to the call is
	// expected
	err = connClient.Call(ctx, "initialized", struct{}{}, nil)
	if err != nil {
		t.Fatalf("failed to send initialized notification: %s", err)
	}

	// 3. wait for the server to load it's config
	timeout := time.NewTimer(defaultTimeout)
	select {
	case <-timeout.C:
		t.Fatalf("timed out waiting for server to load config")
	default:
		for {
			time.Sleep(100 * time.Millisecond)

			if ls.getLoadedConfig() != nil {
				break
			}
		}
	}

	// 4. Touch the new file on disk
	newFilePath := filepath.Join(tempDir, "foo/bar/policy_test.rego")
	newFileURI := uri.FromPath(clients.IdentifierGeneric, newFilePath)
	expectedNewFileURI := uri.FromPath(clients.IdentifierGeneric, filepath.Join(tempDir, "foo/bar_test/policy_test.rego"))

	err = os.MkdirAll(filepath.Dir(newFilePath), 0o755)
	if err != nil {
		t.Fatalf("failed to create directory %s: %s", filepath.Dir(newFilePath), err)
	}

	err = os.WriteFile(newFilePath, []byte(""), 0o600)
	if err != nil {
		t.Fatalf("failed to write file %s: %s", newFilePath, err)
	}

	// 5. Client sends workspace/didCreateFiles notification
	err = connClient.Call(ctx, "workspace/didCreateFiles", types.WorkspaceDidCreateFilesParams{
		Files: []types.WorkspaceDidCreateFilesParamsCreatedFile{
			{URI: newFileURI},
		},
	}, nil)
	if err != nil {
		t.Fatalf("failed to send didChange notification: %s", err)
	}

	// 6. Validate that the client received a workspace edit
	timeout = time.NewTimer(3 * time.Second)
	defer timeout.Stop()

	expectedMessage := fmt.Sprintf(`{
  "edit": {
    "documentChanges": [
      {
        "edits": [
          {
            "newText": "package foo.bar_test\n\nimport rego.v1\n",
            "range": {
              "end": {
                "character": 0,
                "line": 0
              },
              "start": {
                "character": 0,
                "line": 0
              }
            }
          }
        ],
        "textDocument": {
          "uri": "%[1]s",
          "version": null
        }
      },
      {
        "kind": "rename",
        "newUri": "%[2]s",
        "oldUri": "%[1]s",
        "options": {
          "ignoreIfExists": false,
          "overwrite": false
        }
      },
      {
        "kind": "delete",
        "options": {
          "ignoreIfNotExists": true,
          "recursive": true
        },
        "uri": "file://%[3]s/foo/bar"
      }
    ]
  },
  "label": "Template new Rego file"
}`, newFileURI, expectedNewFileURI, tempDir)

	for {
		var success bool
		select {
		case msg := <-receivedMessages:
			t.Log("received message:", string(msg))

			expectedLines := strings.Split(expectedMessage, "\n")
			gotLines := strings.Split(string(msg), "\n")

			if len(gotLines) != len(expectedLines) {
				t.Logf("expected message to have %d lines, got %d", len(expectedLines), len(gotLines))

				continue
			}

			allLinesMatch := true

			for i, expected := range expectedLines {
				if gotLines[i] != expected {
					t.Logf("expected message line %d to be:\n%s\ngot\n%s", i, expected, gotLines[i])

					allLinesMatch = false
				}
			}

			if allLinesMatch {
				success = true
			}
		case <-timeout.C:
			t.Log("never received expected message", expectedMessage)

			t.Fatalf("timed out waiting for expected message to be sent")
		}

		if success {
			break
		}
	}
}
