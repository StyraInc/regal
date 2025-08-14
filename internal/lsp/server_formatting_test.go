package lsp

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/styrainc/regal/internal/lsp/types"
)

func TestFormatting(t *testing.T) {
	t.Parallel()

	// set up the server and client connections
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		t.Fatalf("unexpected request: %v", req)

		return struct{}{}, nil
	}

	tempDir := t.TempDir()
	ls, _ := createAndInitServer(t, ctx, tempDir, clientHandler)
	mainRegoURI := fileURIScheme + filepath.Join(tempDir, "main", "main.rego")

	// Simple as possible — opa fmt should just remove a newline
	ls.cache.SetFileContents(mainRegoURI, "package main\n\n")

	params := types.DocumentFormattingParams{
		TextDocument: types.TextDocumentIdentifier{URI: mainRegoURI},
		Options:      types.FormattingOptions{},
	}

	res, err := ls.handleTextDocumentFormatting(ctx, params)
	if err != nil {
		t.Fatalf("failed to format document: %s", err)
	}

	if edits, ok := res.([]types.TextEdit); ok {
		if len(edits) != 1 {
			t.Fatalf("expected 1 edit, got %d", len(edits))
		}

		expectRange := types.RangeBetween(1, 0, 2, 0)

		if edits[0].Range != expectRange {
			t.Fatalf("expected range to be %v, got %v", expectRange, edits[0].Range)
		}

		if edits[0].NewText != "" {
			t.Fatalf("expected new text to be empty, got %s", edits[0].NewText)
		}
	} else {
		t.Fatalf("expected edits to be []types.TextEdit, got %T", res)
	}
}
