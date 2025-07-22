package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/log"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/uri"
	"github.com/styrainc/regal/internal/testutil"

	"github.com/styrainc/roast/pkg/util/concurrent"
)

func TestTemplateContentsForFile(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		FileKey               string
		CacheFileContents     string
		DiskContents          map[string]string
		RequireConfig         bool
		ServerAllRegoVersions *concurrent.Map[string, ast.RegoVersion]
		ExpectedContents      string
		ExpectedError         string
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
		"empty file is templated based on root": {
			FileKey:           "foo/bar.rego",
			CacheFileContents: "",
			DiskContents: map[string]string{
				"foo/bar.rego":       "",
				".regal/config.yaml": "",
			},
			ExpectedContents: "package foo\n\n",
		},
		"empty test file is templated based on root": {
			FileKey:           "foo/bar_test.rego",
			CacheFileContents: "",
			DiskContents: map[string]string{
				"foo/bar_test.rego":  "",
				".regal/config.yaml": "",
			},
			RequireConfig:    true,
			ExpectedContents: "package foo_test\n\n",
		},
		"empty deeply nested file is templated based on root": {
			FileKey:           "foo/bar/baz/bax.rego",
			CacheFileContents: "",
			DiskContents: map[string]string{
				"foo/bar/baz/bax.rego": "",
				".regal/config.yaml":   "",
			},
			ExpectedContents: "package foo.bar.baz\n\n",
		},
		"v0 templating using rego version setting": {
			FileKey:           "foo/bar/baz/bax.rego",
			CacheFileContents: "",
			ServerAllRegoVersions: concurrent.MapOf(map[string]ast.RegoVersion{
				"foo": ast.RegoV0,
			}),
			DiskContents: map[string]string{
				"foo/bar/baz/bax.rego": "",
				".regal/config.yaml":   "", // we manually set the versions, config not loaded in these tests
			},
			ExpectedContents: "package foo.bar.baz\n\nimport rego.v1\n",
		},
		"v1 templating using rego version setting": {
			FileKey:           "foo/bar/baz/bax.rego",
			CacheFileContents: "",
			ServerAllRegoVersions: concurrent.MapOf(map[string]ast.RegoVersion{
				"foo": ast.RegoV1,
			}),
			DiskContents: map[string]string{
				"foo/bar/baz/bax.rego": "",
				".regal/config.yaml":   "", // we manually set the versions, config not loaded in these tests
			},
			ExpectedContents: "package foo.bar.baz\n\n",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			td := testutil.TempDirectoryOf(t, tc.DiskContents)
			logger := newTestLogger(t)

			ls := NewLanguageServer(
				t.Context(),
				&LanguageServerOptions{LogWriter: logger, LogLevel: log.LevelDebug},
			)

			ls.workspaceRootURI = uri.FromPath(clients.IdentifierGeneric, td)

			ls.loadedConfigAllRegoVersions = tc.ServerAllRegoVersions

			fileURI := uri.FromPath(clients.IdentifierGeneric, filepath.Join(td, tc.FileKey))

			ls.cache.SetFileContents(fileURI, tc.CacheFileContents)

			newContents, err := ls.templateContentsForFile(fileURI)
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

func TestTemplateContentsForFileInWorkspaceRoot(t *testing.T) {
	t.Parallel()

	td := t.TempDir()
	logger := newTestLogger(t)

	testutil.MustMkdirAll(t, td, ".regal")
	testutil.MustWriteFile(t, filepath.Join(td, ".regal", "config.yaml"), []byte{})

	ls := NewLanguageServer(
		t.Context(),
		&LanguageServerOptions{LogWriter: logger, LogLevel: log.LevelDebug},
	)

	ls.workspaceRootURI = uri.FromPath(clients.IdentifierGeneric, td)

	fileURI := uri.FromPath(clients.IdentifierGeneric, filepath.Join(td, "foo.rego"))

	ls.cache.SetFileContents(fileURI, "")

	if _, err := ls.templateContentsForFile(fileURI); err == nil {
		t.Fatalf("expected error")
	} else if !strings.Contains(err.Error(), "this function does not template files in the workspace root") {
		t.Fatalf("expected error about root templating, got %s", err.Error())
	}
}

func TestTemplateContentsForFileWithUnknownRoot(t *testing.T) {
	t.Parallel()

	td := t.TempDir()
	logger := newTestLogger(t)

	ls := NewLanguageServer(
		t.Context(),
		&LanguageServerOptions{LogWriter: logger, LogLevel: log.LevelDebug},
	)

	ls.workspaceRootURI = uri.FromPath(clients.IdentifierGeneric, td)

	testutil.MustMkdirAll(t, td, "foo")

	fileURI := uri.FromPath(clients.IdentifierGeneric, filepath.Join(td, "foo", "bar.rego"))

	ls.cache.SetFileContents(fileURI, "")

	newContents, err := ls.templateContentsForFile(fileURI)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	exp := "package foo\n\n"

	if exp != newContents {
		t.Errorf("unexpected content: %s, want %s", newContents, exp)
	}
}

func TestNewFileTemplating(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		".regal/config.yaml": `rules:
  idiomatic:
    directory-package-mismatch:
      level: error
      exclude-test-suffix: false
`,
	}

	tempDir := testutil.TempDirectoryOf(t, files)

	// set up the server and client connections
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	receivedMessages := make(chan []byte, 10)
	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		bs, err := json.MarshalIndent(req.Params, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal params: %s", err)
		}

		receivedMessages <- bs

		return struct{}{}, nil
	}

	ls, connClient := createAndInitServer(t, ctx, newTestLogger(t), tempDir, clientHandler)

	go ls.StartTemplateWorker(ctx)

	// wait for the server to load it's config
	timeout := time.NewTimer(determineTimeout())
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

	// Touch the new file on disk
	newFilePath := filepath.Join(tempDir, "foo", "bar", "policy_test.rego")
	newFileURI := uri.FromPath(clients.IdentifierGeneric, newFilePath)
	expectedNewFileURI := uri.FromPath(clients.IdentifierGeneric, filepath.Join(
		tempDir, "foo", "bar_test", "policy_test.rego",
	))

	testutil.MustMkdirAll(t, filepath.Dir(newFilePath))
	testutil.MustWriteFile(t, newFilePath, []byte(""))

	// Client sends workspace/didCreateFiles notification
	if err := connClient.Notify(ctx, "workspace/didCreateFiles", types.WorkspaceDidCreateFilesParams{
		Files: []types.WorkspaceDidCreateFilesParamsCreatedFile{
			{URI: newFileURI},
		},
	}, nil); err != nil {
		t.Fatalf("failed to send didChange notification: %s", err)
	}

	// Validate that the client received a workspace edit
	timeout.Reset(determineTimeout())

	expectedMessage := fmt.Sprintf(`{
  "edit": {
    "documentChanges": [
      {
        "edits": [
          {
            "newText": "package foo.bar_test\n\n",
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

	for success := false; !success; {
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

			success = allLinesMatch
		case <-timeout.C:
			t.Log("never received expected message", expectedMessage)

			t.Fatalf("timed out waiting for expected message to be sent")
		}
	}
}

// TestTemplateWorkerRaceConditionWithDidOpen tests the race condition fix for
// https://github.com/StyraInc/regal/issues/1608 where didOpen would overwrite
// templated content in cache (with "") before the template worker could
// complete.
func TestTemplateWorkerRaceConditionWithDidOpen(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		".regal/config.yaml": `{}`,
	}

	tempDir := testutil.TempDirectoryOf(t, files)

	// set up the server and client connections
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	receivedMessages := make(chan []byte, 10)
	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		bs, err := json.MarshalIndent(req.Params, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal params: %s", err)
		}

		receivedMessages <- bs

		return struct{}{}, nil
	}

	ls, connClient := createAndInitServer(t, ctx, newTestLogger(t), tempDir, clientHandler)

	newFilePath := filepath.Join(tempDir, "foo", "bar", "policy.rego")
	newFileURI := uri.FromPath(clients.IdentifierGeneric, newFilePath)

	testutil.MustMkdirAll(t, filepath.Dir(newFilePath))
	testutil.MustWriteFile(t, newFilePath, []byte(""))

	// create a 'manual' worker so we can time events to recreated the race.
	templateCompleted := make(chan bool, 1)
	proceedWithTemplating := make(chan bool, 1)

	go func() {
		job := <-ls.templateFileJobs
		t.Log("controlled template worker received job:", job.URI)

		ls.templatingFiles.Set(job.URI, true)
		t.Log("marked file as being templated")

		<-proceedWithTemplating
		t.Log("proceeding with templating...")

		templateContent := "package foo.bar\n\n"
		ls.cache.SetFileContents(job.URI, templateContent)
		t.Log("set templated content in cache")

		ls.templatingFiles.Delete(job.URI)
		t.Log("template worker completed - cleared templating flag")

		templateCompleted <- true
	}()

	// 1: send didCreateFiles to trigger templating
	if err := connClient.Notify(ctx, "workspace/didCreateFiles", types.WorkspaceDidCreateFilesParams{
		Files: []types.WorkspaceDidCreateFilesParamsCreatedFile{
			{URI: newFileURI},
		},
	}, nil); err != nil {
		t.Fatalf("failed to send didCreateFiles notification: %s", err)
	}

	// wait for lock to be set, the lock is set quickly, but we want reliable tests
	// so we make sure that the lock is set before proceeding
	timeout := time.NewTimer(determineTimeout())
	select {
	case <-timeout.C:
		t.Fatalf("timed out waiting for lock to  be set")
	default:
		for {
			isLocked, ok := ls.templatingFiles.Get(newFileURI)
			if !ok {
				t.Log("expected file to be locked for templating")
			}

			if isLocked {
				break
			}
		}
	}

	// 2: send the didOpen notification while templating is in progress
	if err := connClient.Notify(ctx, "textDocument/didOpen", types.TextDocumentDidOpenParams{
		TextDocument: types.TextDocumentItem{
			URI:        newFileURI,
			LanguageID: "rego",
			Version:    1,
			Text:       "", // the file is empty in the editor
		},
	}, nil); err != nil {
		t.Fatalf("failed to send didOpen notification: %s", err)
	}

	// 3: allow templating to continue.
	proceedWithTemplating <- true

	<-templateCompleted

	cacheContent, ok := ls.cache.GetFileContents(newFileURI)
	if !ok {
		t.Fatalf("expected file to be in cache")
	}

	expectedTemplateContent := "package foo.bar\n\n"
	if cacheContent != expectedTemplateContent {
		t.Fatalf("race condition occurred! Expected cache to contain %q, got %q. "+
			"didOpen overwrote templated content.", expectedTemplateContent, cacheContent)
	}

	// 4: check that the didOpen now works again after the lock is off
	if err := connClient.Notify(ctx, "textDocument/didOpen", types.TextDocumentDidOpenParams{
		TextDocument: types.TextDocumentItem{
			URI:        newFileURI,
			LanguageID: "rego",
			Version:    2,
			Text:       "package foo.bar\n\nimport rego.v1\n", // file has some content now
		},
	}, nil); err != nil {
		t.Fatalf("failed to send second didOpen notification: %s", err)
	}

	// wait for contents to be set
	expectedFinalContent := "package foo.bar\n\nimport rego.v1\n"
	timeout = time.NewTimer(determineTimeout())
	select {
	case <-timeout.C:
		t.Fatalf("timed out waiting for did open content to be set")
	default:
		for {
			time.Sleep(100 * time.Millisecond)

			finalContent, ok := ls.cache.GetFileContents(newFileURI)
			if !ok {
				t.Log("expected file to be in cache after second didOpen")
			}

			if finalContent == expectedFinalContent {
				break
			}

			t.Logf("expected cache to be updated to %q, got %q", expectedFinalContent, finalContent)
		}
	}
}
