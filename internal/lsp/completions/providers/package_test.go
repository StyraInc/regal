package providers

import (
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
)

func TestPackage(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := "\n"

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &Package{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      0,
			Character: 0,
		},
	}

	completions, err := p.Run(c, completionParams)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(completions) != 1 {
		t.Fatalf("Expected exactly one completion, got: %v", completions)
	}

	comp := completions[0]
	if comp.Label != "package" {
		t.Fatalf("Expected label to be 'package', got: %v", comp.Label)
	}
}

func TestPackageAfterComment(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `
# this is a comment before the package statement
p

`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &Package{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 1,
		},
	}

	completions, err := p.Run(c, completionParams)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(completions) != 1 {
		t.Fatalf("Expected exactly one completion, got: %v", completions)
	}

	comp := completions[0]
	if comp.Label != "package" {
		t.Fatalf("Expected label to be 'package', got: %v", comp.Label)
	}
}

func TestPackageNotLaterLines(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := "package foo\n\n"

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &Package{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      1,
			Character: 0,
		},
	}

	completions, err := p.Run(c, completionParams)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(completions) != 0 {
		t.Fatalf("Expected no completions, got: %v", completions)
	}
}
