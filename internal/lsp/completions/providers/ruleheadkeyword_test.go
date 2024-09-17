//nolint:dupl
package providers

import (
	"context"
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
)

func TestRuleHeadKeyword_ManualTriggerAfterRuleName(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package policy

deny` + " "

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &RuleHeadKeyword{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 4,
		},
	}

	completions, err := p.Run(context.Background(), c, completionParams, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedLabels := []string{"contains", "if", ":="}

	if len(completions) != len(expectedLabels) {
		t.Fatalf("Expected %v completions, got: %v", len(expectedLabels), len(completions))
	}

	for _, comp := range completions {
		found := false

		for _, label := range expectedLabels {
			if comp.Label == label {
				found = true

				break
			}
		}

		if !found {
			t.Fatalf("Expected label to be one of %v, got: %v", expectedLabels, comp.Label)
		}
	}
}

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

	completions, err := p.Run(context.Background(), c, completionParams, nil)
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

	completions, err := p.Run(context.Background(), c, completionParams, nil)
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

	completions, err := p.Run(context.Background(), c, completionParams, nil)
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
