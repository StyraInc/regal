package lsp

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/styrainc/regal/internal/lsp/types"

	"github.com/styrainc/roast/pkg/encoding"
)

// TestSendFileDiagnosticsEmptyArrays replicates the scenario from
// https://github.com/StyraInc/regal/issues/1609 where a file that's been
// deleted from the cache has null rather than empty arrays as diagnostics.
func TestSendFileDiagnosticsEmptyArrays(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parseErrors         []types.Diagnostic
		lintErrors          []types.Diagnostic
		fileInCache         bool
		expectedDiagnostics []types.Diagnostic
	}{
		"lint errors only": {
			parseErrors:         []types.Diagnostic{},
			lintErrors:          []types.Diagnostic{{Message: "lint error"}},
			fileInCache:         true,
			expectedDiagnostics: []types.Diagnostic{{Message: "lint error"}},
		},
		"parse errors only": {
			parseErrors:         []types.Diagnostic{{Message: "parse error"}},
			lintErrors:          []types.Diagnostic{},
			fileInCache:         true,
			expectedDiagnostics: []types.Diagnostic{{Message: "parse error"}},
		},
		"both empty in cache": {
			parseErrors:         []types.Diagnostic{},
			lintErrors:          []types.Diagnostic{},
			fileInCache:         true,
			expectedDiagnostics: []types.Diagnostic{},
		},
		"file deleted from cache": {
			// file deleted, and so nothing in the cache
			fileInCache:         false,
			expectedDiagnostics: []types.Diagnostic{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			fileURI := "file:///test.rego"
			receivedDiagnostics := make(chan types.FileDiagnostics, 1)

			clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
				if req.Method != methodTextDocumentPublishDiagnostics {
					t.Fatalf("expected method %s, got %s", methodTextDocumentPublishDiagnostics, req.Method)
				}

				var requestData types.FileDiagnostics
				if err := encoding.JSON().Unmarshal(*req.Params, &requestData); err != nil {
					t.Fatalf("failed to unmarshal diagnostics: %s", err)
				}

				receivedDiagnostics <- requestData

				return struct{}{}, nil
			}

			ls, _ := createAndInitServer(t, ctx, newTestLogger(t), t.TempDir(), clientHandler)

			if tc.fileInCache {
				ls.cache.SetParseErrors(fileURI, tc.parseErrors)
				ls.cache.SetFileDiagnostics(fileURI, tc.lintErrors)
			}

			if err := ls.sendFileDiagnostics(ctx, fileURI); err != nil {
				t.Fatalf("sendFileDiagnostics failed: %v", err)
			}

			select {
			case diag := <-receivedDiagnostics:
				if diag.URI != fileURI {
					t.Fatalf("expected URI %s, got %s", fileURI, diag.URI)
				}

				if len(tc.expectedDiagnostics) == 0 && diag.Items == nil {
					t.Errorf("expected empty array [], got nil")
				}

				if len(diag.Items) != len(tc.expectedDiagnostics) {
					t.Errorf("expected %d diagnostics, got %d", len(tc.expectedDiagnostics), len(diag.Items))
				}

				for i, expected := range tc.expectedDiagnostics {
					if i < len(diag.Items) && diag.Items[i].Message != expected.Message {
						t.Errorf("diagnostic %d: expected message %s, got %s", i, expected.Message, diag.Items[i].Message)
					}
				}

			case <-time.After(100 * time.Millisecond):
				t.Fatal("no diagnostics received before timeout")
			}
		})
	}
}
