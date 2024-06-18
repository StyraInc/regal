package completions

import (
	"testing"

	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/completions/providers"
	"github.com/styrainc/regal/internal/lsp/types"
)

func TestManager(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()
	fileURI := "file:///foo/bar/file.rego"
	fileContents := ""

	c.SetFileContents(fileURI, fileContents)

	mgr := NewManager(c, &ManagerOptions{})
	mgr.RegisterProvider(&providers.Package{})

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: fileURI,
		},
		Position: types.Position{
			Line:      0,
			Character: 1,
		},
	}

	completions, err := mgr.Run(completionParams, nil)
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

func TestManagerEarlyExitInsideComment(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()
	fileURI := "file:///foo/bar/file.rego"

	fileContents := `package p

# foo := http
`

	module := ast.MustParseModule(fileContents)

	c.SetFileContents(fileURI, fileContents)
	c.SetModule(fileURI, module)

	mgr := NewManager(c, &ManagerOptions{})
	mgr.RegisterProvider(&providers.BuiltIns{})

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: fileURI,
		},
		Position: types.Position{
			Line:      2,
			Character: 13,
		},
	}

	completions, err := mgr.Run(completionParams, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(completions) != 0 {
		t.Errorf("Expected no completions, got: %v", completions)
	}
}
