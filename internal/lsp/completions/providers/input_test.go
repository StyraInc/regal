package providers

import (
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
)

func TestInput_if(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package foo

allow if i`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &Input{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 10, // is the c char that triggered the request
		},
	}

	completions, err := p.Run(c, completionParams, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	labels := completionLabels(completions)

	if len(labels) != 1 {
		t.Fatalf("Expected one completion, got: %v", labels)
	}

	if labels[0] != "input" {
		t.Fatalf("Expected 'input' completion, got: %v", labels[0])
	}

	if exp, got := "input", completions[0].TextEdit.NewText; exp != got {
		t.Fatalf("Expected '%s' as new text, got: %s", exp, got)
	}
}
