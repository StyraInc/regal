package providers

import (
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
)

func TestRegoV1(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileURI := "file:///foo/bar/file.rego"
	fileContents := `package fo

import r

`

	c.SetFileContents(fileURI, fileContents)

	p := &RegoV1{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: fileURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 8,
		},
	}

	completions, err := p.Run(c, completionParams, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(completions) != 1 {
		t.Fatalf("Expected exactly one completion, got: %v", completions)
	}

	comp := completions[0]
	if comp.Label != "rego.v1" {
		t.Fatalf("Expected label to be 'package', got: %v", comp.Label)
	}
}
