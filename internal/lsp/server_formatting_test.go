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

	tempDir := t.TempDir()

	// set up the server and client connections
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientHandler := func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
		t.Fatalf("unexpected request: %v", req)

		return struct{}{}, nil
	}

	ls, _ := createAndInitServer(t, ctx, newTestLogger(t), tempDir, clientHandler)
	mainRegoURI := fileURIScheme + filepath.Join(tempDir, "main", "main.rego")

	// Simple as possible — opa fmt should just remove a newline
	content := `package main

`
	ls.cache.SetFileContents(mainRegoURI, content)

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

		expectRange := types.Range{
			Start: types.Position{Line: 1, Character: 0},
			End:   types.Position{Line: 2, Character: 0},
		}

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
