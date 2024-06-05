//nolint:dupl
package providers

import (
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/completions/refs"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/parse"
)

func TestCommonRule_TypedA(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package policy

a
`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &CommonRule{}

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
	if comp.Label != "allow" {
		t.Fatalf("Expected label to be 'allow', got: %v", comp.Label)
	}
}

func TestCommonRule_TypedD(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package policy

d
`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &CommonRule{}

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
	if comp.Label != "deny" {
		t.Fatalf("Expected label to be 'deny', got: %v", comp.Label)
	}
}

func TestCommonRule_TypedDAlreadyDefined(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package policy

deny := false

`

	c.SetFileContents(testCaseFileURI, fileContents+"d")

	mod, err := parse.Module(testCaseFileURI, fileContents)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	c.SetModule(testCaseFileURI, mod)
	c.SetFileRefs(testCaseFileURI, refs.DefinedInModule(mod))

	p := &CommonRule{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      4,
			Character: 1,
		},
	}

	completions, err := p.Run(c, completionParams, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(completions) != 0 {
		t.Fatalf("Expected no completions, got: %v", completions)
	}
}
