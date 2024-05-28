//nolint:dupl
package providers

import (
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/completion"
)

func TestDefaultInvoked(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := "package policy\n\n"

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &Default{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      3,
			Character: 0,
		},
		Context: types.CompletionContext{
			TriggerKind: completion.Invoked,
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
	if comp.Label != "default" {
		t.Fatalf("Expected label to be 'default', got: %v", comp.Label)
	}
}

func TestDefaultTypedD(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := "package policy\n\nd"

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &Default{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 1,
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
	if comp.Label != "default" {
		t.Fatalf("Expected label to be 'default', got: %v", comp.Label)
	}
}
