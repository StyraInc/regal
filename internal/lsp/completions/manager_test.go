package completions

import (
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/storage/inmem"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/completions/providers"
	"github.com/styrainc/regal/internal/lsp/types"
)

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

func TestManagerRankCompletions(t *testing.T) {
	t.Parallel()

	file1Contents := `package example

import rego.v1

foo := true
`

	file1 := ast.MustParseModule(file1Contents)

	file2Contents := `package example2

import rego.v1

bar := data.example.foo
`

	file2 := ast.MustParseModule(file2Contents)

	store := inmem.NewFromObject(map[string]interface{}{
		"workspace": map[string]interface{}{
			"parsed": map[string]interface{}{
				"file:///file1.rego": file1,
				"file:///file2.rego": file2,
			},
		},
	})

	policyProvider := providers.NewPolicy(store)

	file2ContentsEdited := `package example2

import rego.v1

bar := data.example.foo

baz := data.
`

	c := cache.NewCache()

	c.SetFileContents("file:///file2.rego", file2ContentsEdited)
	c.SetUsedRefs("file:///file2.rego", []string{"data.example.foo"})

	mgr := NewManager(c, &ManagerOptions{})
	mgr.RegisterProvider(&providers.UsedRefs{})
	mgr.RegisterProvider(policyProvider)

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: "file:///file2.rego",
		},
		Position: types.Position{
			Line:      6,
			Character: 13,
		},
	}

	completions, err := mgr.Run(completionParams, &providers.Options{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	count := 0

	for _, item := range completions {
		if item.Label == "data.example.foo" {
			count++
		}
	}

	if count != 1 {
		t.Fatalf("Expected exactly one completion for 'data.example.foo', got: %v", completions)
	}
}
