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
)

// TestLanguageServerSingleFile tests that changes to a single file and Regal config are handled correctly by the
// language server by making updates to both and validating that the correct diagnostics are sent to the client.
//
// This test also ensures that updating the config to point to a non-default engine and capabilities version works
// and causes that engine's builtins to work with completions.
//
//nolint:maintidx
func TestLanguageServerSingleFile(t *testing.T) {
	t.Parallel()

	// set up the workspace content with some example rego and regal config
	tempDir := t.TempDir()
	mainRegoURI := fileURIScheme + tempDir + mainRegoFileName

	if err := os.MkdirAll(filepath.Join(tempDir, ".regal"), 0o755); err != nil {
		t.Fatalf("failed to create .regal directory: %s", err)
	}

	mainRegoContents := `package main

import rego.v1
allow = true
`

	files := map[string]string{
		"main.rego": mainRegoContents,
		".regal/config.yaml": `
rules:
  idiomatic:
    directory-package-mismatch:
      level: ignore`,
	}

	for f, fc := range files {
		if err := os.WriteFile(filepath.Join(tempDir, f), []byte(fc), 0o600); err != nil {
			t.Fatalf("failed to write file %s: %s", f, err)
		}
	}

	receivedMessages := make(chan types.FileDiagnostics, defaultBufferedChannelSize)
	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		if req.Method == methodTextDocumentPublishDiagnostics {
			var requestData types.FileDiagnostics

			if err := encoding.JSON().Unmarshal(*req.Params, &requestData); err != nil {
				t.Fatalf("failed to unmarshal diagnostics: %s", err)
			}

			receivedMessages <- requestData

			return struct{}{}, nil
		}

		t.Fatalf("unexpected request: %v", req)

		return struct{}{}, nil
	}

	// set up the server and client connections
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, connClient, err := createAndInitServer(ctx, newTestLogger(t), tempDir, files, clientHandler)
	if err != nil {
		t.Fatalf("failed to create and init language server: %s", err)
	}

	// validate that the client received a diagnostics notification for the file
	timeout := time.NewTimer(determineTimeout())
	defer timeout.Stop()

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			success = testRequestDataCodes(t, requestData, mainRegoURI, []string{"opa-fmt", "use-assignment-operator"})
		case <-timeout.C:
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// Client sends textDocument/didChange notification with new contents for main.rego
	// no response to the call is expected
	if err := connClient.Call(ctx, "textDocument/didChange", types.TextDocumentDidChangeParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: mainRegoURI,
		},
		ContentChanges: []types.TextDocumentContentChangeEvent{
			{
				Text: `package main
import rego.v1
allow := true
`,
			},
		},
	}, nil); err != nil {
		t.Fatalf("failed to send didChange notification: %s", err)
	}

	// validate that the client received a new diagnostics notification for the file
	timeout.Reset(determineTimeout())

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			success = testRequestDataCodes(t, requestData, mainRegoURI, []string{"opa-fmt"})
		case <-timeout.C:
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// config update is caught by the config watcher
	newConfigContents := `
rules:
  idiomatic:
    directory-package-mismatch:
      level: ignore
  style:
    opa-fmt:
      level: ignore
`

	if err := os.WriteFile(filepath.Join(tempDir, ".regal/config.yaml"), []byte(newConfigContents), 0o600); err != nil {
		t.Fatalf("failed to write new config file: %s", err)
	}

	// validate that the client received a new, empty diagnostics notification for the file
	timeout.Reset(determineTimeout())

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			if requestData.URI != mainRegoURI {
				t.Logf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)

				continue
			}

			codes := []string{}
			for _, d := range requestData.Items {
				codes = append(codes, d.Code)
			}

			if len(requestData.Items) != 0 {
				t.Logf("expected empty diagnostics, got %v", codes)

				continue
			}

			success = testRequestDataCodes(t, requestData, mainRegoURI, []string{})
		case <-timeout.C:
			t.Fatalf("timed out waiting for main.rego diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// Client sends new config with an EOPA capabilities file specified.
	newConfigContents = `
rules:
  style:
    opa-fmt:
      level: ignore
  idiomatic:
    directory-package-mismatch:
      level: ignore
capabilities:
  from:
    engine: eopa
    version: v1.23.0
`

	if err := os.WriteFile(filepath.Join(tempDir, ".regal/config.yaml"), []byte(newConfigContents), 0o600); err != nil {
		t.Fatalf("failed to write new config file: %s", err)
	}

	// validate that the client received a new, empty diagnostics notification for the file
	timeout.Reset(determineTimeout())

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			if requestData.URI != mainRegoURI {
				t.Logf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)

				break
			}

			codes := []string{}
			for _, d := range requestData.Items {
				codes = append(codes, d.Code)
			}

			if len(requestData.Items) != 0 {
				t.Logf("expected empty diagnostics, got %v", codes)

				continue
			}

			success = testRequestDataCodes(t, requestData, mainRegoURI, []string{})
		case <-timeout.C:
			t.Fatalf("timed out waiting for main.rego diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// Client sends textDocument/didChange notification with new
	// contents for main.rego no response to the call is expected. We added
	// the start of an EOPA-specific call, so if the capabilities were
	// loaded correctly, we should see a completion later after we ask for
	// it.
	if err := connClient.Call(ctx, "textDocument/didChange", types.TextDocumentDidChangeParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: mainRegoURI,
		},
		ContentChanges: []types.TextDocumentContentChangeEvent{
			{
				Text: `package main
import rego.v1

# METADATA
# entrypoint: true
allow := neo4j.q
`,
			},
		},
	}, nil); err != nil {
		t.Fatalf("failed to send didChange notification: %s", err)
	}

	// validate that the client received a new diagnostics notification for the file
	timeout.Reset(determineTimeout())

	for {
		var success bool
		select {
		case requestData := <-receivedMessages:
			if requestData.URI != mainRegoURI {
				t.Logf("expected diagnostics to be sent for main.rego, got %s", requestData.URI)

				break
			}

			codes := []string{}
			for _, d := range requestData.Items {
				codes = append(codes, d.Code)
			}

			if len(requestData.Items) != 0 {
				t.Logf("expected empty diagnostics, got %v", codes)

				continue
			}

			success = testRequestDataCodes(t, requestData, mainRegoURI, []string{})
		case <-timeout.C:
			t.Fatalf("timed out waiting for file diagnostics to be sent")
		}

		if success {
			break
		}
	}

	// 7. With our new config applied, and the file updated, we can ask the
	// LSP for a completion. We expect to see neo4j.query show up. Since
	// neo4j.query is an EOPA-specific builtin, it should never appear if
	// we're using the normal OPA capabilities file.
	timeout.Reset(determineTimeout())

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		foundNeo4j := false

		select {
		case <-ticker.C:
			// Create a new context with timeout for each request, this is
			// timed out after 1s as GHA runner sometimes takes a while.
			reqCtx, reqCtxCancel := context.WithTimeout(ctx, time.Second)

			resp := make(map[string]any)
			err := connClient.Call(reqCtx, "textDocument/completion", types.CompletionParams{
				TextDocument: types.TextDocumentIdentifier{
					URI: mainRegoURI,
				},
				Position: types.Position{
					Line:      5,
					Character: 16,
				},
			}, &resp)

			reqCtxCancel()

			if err != nil {
				t.Fatalf("failed to send completion request: %s", err)
			}

			itemsList, ok := resp["items"].([]any)
			if !ok {
				t.Fatalf("failed to cast resp[items] to []any")
			}

			for _, itemI := range itemsList {
				item, ok := itemI.(map[string]any)
				if !ok {
					t.Fatalf("completion item '%+v' was not a JSON object", itemI)
				}

				label, ok := item["label"].(string)
				if !ok {
					t.Fatalf("completion item label is not a string: %+v", item["label"])
				}

				if label == "neo4j.query" {
					foundNeo4j = true

					break
				}
			}

			t.Logf("waiting for neo4j.query in completion results for neo4j.q, got %v", itemsList)
		case <-timeout.C:
			t.Fatalf("timed out waiting for file completion to correct")
		}

		if foundNeo4j {
			break
		}
	}
}
