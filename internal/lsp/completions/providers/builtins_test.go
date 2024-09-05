package providers

import (
	"slices"
	"strings"
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
)

func TestBuiltIns_if(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package foo

allow if c`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &BuiltIns{}

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

	if !slices.Contains(labels, "count") {
		t.Fatalf("Expected to find 'count' in completions, got: %s", strings.Join(labels, ", "))
	}
}

func TestBuiltIns_afterAssignment(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package foo

allow := c`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &BuiltIns{}

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

	if !slices.Contains(labels, "count") {
		t.Fatalf("Expected to find 'count' in completions, got: %s", strings.Join(labels, ", "))
	}
}

func TestBuiltIns_inRuleBody(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package foo

allow if {
  c
}`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &BuiltIns{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      3,
			Character: 3, // is the c char that triggered the request
		},
	}

	completions, err := p.Run(c, completionParams, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	labels := completionLabels(completions)

	if !slices.Contains(labels, "count") {
		t.Fatalf("Expected to find 'count' in completions, got: %s", strings.Join(labels, ", "))
	}
}

func TestBuiltIns_noInfix(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package foo

allow if gt`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &BuiltIns{}

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

	if len(completions) != 0 {
		t.Fatalf("Expected no completions, got: %v", completions)
	}
}

func TestBuiltIns_noDeprecated(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package foo

allow if c`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &BuiltIns{}

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

	if slices.Contains(labels, "cast_set") {
		t.Fatalf("Expected no deprecated completions, got: %s", strings.Join(labels, ", "))
	}
}

func TestBuiltIns_noDefaultRule(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := `package foo

default allow := f`

	c.SetFileContents(testCaseFileURI, fileContents)

	p := &BuiltIns{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 18, // is the c char that triggered the request
		},
	}

	completions, err := p.Run(c, completionParams, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(completions) != 0 {
		t.Fatalf("Expected no completions, got: %d", len(completions))
	}
}

func completionLabels(completions []types.CompletionItem) []string {
	labels := make([]string, len(completions))
	for i, c := range completions {
		labels[i] = c.Label
	}

	return labels
}
