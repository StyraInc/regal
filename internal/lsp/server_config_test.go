package lsp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anderseknert/roast/pkg/encoding"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/uri"
)

// TestLanguageServerParentDirConfig tests that regal config is loaded as it is for the
// Regal CLI, and that config files in a parent directory are loaded correctly
// even when the workspace is a child directory.
func TestLanguageServerParentDirConfig(t *testing.T) {
	t.Parallel()

	var err error

	// this is the top level directory for the test
	tempDir := t.TempDir()
	// childDir will be the directory that the client is using as its workspace

	childDirName := "child"
	childDir := filepath.Join(tempDir, childDirName)

	mainRegoContents := `package main

import rego.v1
allow := true
`

	files := map[string]string{
		childDirName + mainRegoFileName: mainRegoContents,
		".regal/config.yaml": `rules:
  idiomatic:
    directory-package-mismatch:
      level: ignore
  style:
    opa-fmt:
      level: error
`,
	}

	// mainRegoFileURI is used throughout the test to refer to the main.rego file
	// and so it is defined here for convenience
	mainRegoFileURI := fileURIScheme + childDir + mainRegoFileName

	// set up the server and client connections
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	receivedMessages := make(chan types.FileDiagnostics, defaultBufferedChannelSize)
	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		if req.Method == methodTextDocumentPublishDiagnostics {
			var requestData types.FileDiagnostics

			if err2 := encoding.JSON().Unmarshal(*req.Params, &requestData); err2 != nil {
				t.Fatalf("failed to unmarshal diagnostics: %s", err2)
			}

			receivedMessages <- requestData

			return struct{}{}, nil
		}

		t.Logf("unexpected request from server: %v", req)

		return struct{}{}, nil
	}

	ls, _, err := createAndInitServer(ctx, newTestLogger(t), tempDir, files, clientHandler)
	if err != nil {
		t.Fatalf("failed to create and init language server: %s", err)
	}

	if got, exp := ls.workspaceRootURI, uri.FromPath(ls.clientIdentifier, tempDir); exp != got {
		t.Fatalf("expected client root URI to be %s, got %s", exp, got)
	}

	timeout := time.NewTimer(defaultTimeout)
	defer timeout.Stop()

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			success = testRequestDataCodes(t, requestData, mainRegoFileURI, []string{"opa-fmt"})
		case <-timeout.C:
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// User updates config file contents in parent directory that is not
	// part of the workspace
	newConfigContents := `rules:
  idiomatic:
    directory-package-mismatch:
      level: ignore
  style:
    opa-fmt:
      level: ignore
`

	path := filepath.Join(tempDir, ".regal/config.yaml")
	if err := os.WriteFile(path, []byte(newConfigContents), 0o600); err != nil {
		t.Fatalf("failed to write new config file: %s", err)
	}

	// validate that the client received a new, empty diagnostics notification for the file
	timeout.Reset(defaultTimeout)

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			success = testRequestDataCodes(t, requestData, mainRegoFileURI, []string{})
		case <-timeout.C:
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
	}
}
