//nolint:dupl
package providers

import (
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
)

func TestRuleHeadKeyword_TypedIAfterRuleName(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package policy

deny i
`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &RuleHeadKeyword{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 5,
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
	if comp.Label != "if" {
		t.Fatalf("Expected label to be 'if', got: %v", comp.Label)
	}

	if comp.Mandatory != true {
		t.Fatalf("Expected mandatory to be true, got: %v", comp.Mandatory)
	}
}

func TestRuleHeadKeyword_TypedC(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package policy

deny c
`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &RuleHeadKeyword{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 5,
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
	if comp.Label != "contains" {
		t.Fatalf("Expected label to be 'contains', got: %v", comp.Label)
	}

	if comp.Mandatory != true {
		t.Fatalf("Expected mandatory to be true, got: %v", comp.Mandatory)
	}
}

func TestRuleHeadKeyword_TypedI(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package policy

deny contains message i
`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &RuleHeadKeyword{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 23,
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
	if comp.Label != "if" {
		t.Fatalf("Expected label to be 'if' got: %v", comp.Label)
	}

	if comp.Mandatory != true {
		t.Fatalf("Expected mandatory to be true, got: %v", comp.Mandatory)
	}
}
