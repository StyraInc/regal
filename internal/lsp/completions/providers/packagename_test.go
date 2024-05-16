package providers

import (
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
)

func TestPackageName(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileContents := "package "

	fileURI := "file:///foo/bar/baz/bax/file.rego"

	c.SetFileContents(fileURI, fileContents)

	p := &PackageName{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: fileURI,
		},
		Position: types.Position{
			Line:      0,
			Character: 9,
		},
	}

	completions, err := p.Run(c, completionParams, &Options{
		RootURI: "file:///foo/bar",
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(completions) != 1 {
		t.Fatalf("Expected exactly one completion, got: %v", completions)
	}

	comp := completions[0]
	if comp.Label != "package baz.bax" {
		t.Fatalf("Expected label to be 'baz.bax', got: %v", comp.Label)
	}
}

func TestPackageNameWithPackageComment(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileURI := "file:///bar/foo/file.rego"
	fileContents := `
# this is a comment before the package statement
# at the start of a file

package `

	c.SetFileContents(fileURI, fileContents)

	p := &PackageName{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: fileURI,
		},
		Position: types.Position{
			Line:      4,
			Character: 9,
		},
	}

	completions, err := p.Run(c, completionParams, &Options{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(completions) != 1 {
		t.Fatalf("Expected exactly one completion, got: %v", completions)
	}

	comp := completions[0]
	if comp.Label != "package foo" {
		t.Fatalf("Expected label to be 'foo', got: %v", comp.Label)
	}
}

func TestPackageNameWithErroneousPackageStatements(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	fileURI := "file:///foo/bar/file.rego"
	fileContents := `package foo

package `

	c.SetFileContents(fileURI, fileContents)

	p := &PackageName{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: fileURI,
		},
		Position: types.Position{
			Line:      4,
			Character: 9,
		},
	}

	completions, err := p.Run(c, completionParams, &Options{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(completions) != 0 {
		t.Fatalf("Expected no completions, got: %v", completions)
	}
}
