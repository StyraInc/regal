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

	fileURI := "file:///foo/bar/file.rego"
	fileContents := `package foo

allow if c`

	c.SetFileContents(fileURI, fileContents)

	p := &BuiltIns{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: fileURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 10, // is the c char that triggered the request
		},
	}

	completions, err := p.Run(c, completionParams)
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

	fileURI := "file:///foo/bar/file.rego"
	fileContents := `package foo

allow := c`

	c.SetFileContents(fileURI, fileContents)

	p := &BuiltIns{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: fileURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 10, // is the c char that triggered the request
		},
	}

	completions, err := p.Run(c, completionParams)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	labels := completionLabels(completions)

	if !slices.Contains(labels, "count") {
		t.Fatalf("Expected to find 'count' in completions, got: %s", strings.Join(labels, ", "))
	}
}

func TestBuiltIns_inRuleBody(t *testing.T) {
	c := cache.NewCache()

	fileURI := "file:///foo/bar/file.rego"
	fileContents := `package foo

allow if {
  c
}`

	c.SetFileContents(fileURI, fileContents)

	p := &BuiltIns{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: fileURI,
		},
		Position: types.Position{
			Line:      3,
			Character: 3, // is the c char that triggered the request
		},
	}

	completions, err := p.Run(c, completionParams)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	labels := completionLabels(completions)

	if !slices.Contains(labels, "count") {
		t.Fatalf("Expected to find 'count' in completions, got: %s", strings.Join(labels, ", "))
	}
}

func completionLabels(completions []types.CompletionItem) []string {
	labels := make([]string, len(completions))
	for i, c := range completions {
		labels[i] = c.Label
	}
	return labels
}
