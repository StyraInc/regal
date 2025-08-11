package providers

import (
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/storage/inmem"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/test"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/pkg/roast/encoding"
)

func TestPolicyProvider_Example1(t *testing.T) {
	t.Parallel()

	policy := `package p

allow if {
	user := data.users[0]
	# try completion on next line
	roles := u
}
`
	module := parse.MustParseModule(policy)
	moduleMap := make(map[string]any)

	encoding.MustJSONRoundTrip(module, &moduleMap)

	c := cache.NewCache()
	c.SetFileContents(testCaseFileURI, policy)

	store := inmem.NewFromObjectWithOpts(map[string]any{
		"workspace": map[string]any{
			"parsed": map[string]any{
				testCaseFileURI: moduleMap,
			},
		},
	}, inmem.OptRoundTripOnWrite(false))

	params := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{URI: testCaseFileURI},
		Position:     types.Position{Line: 5, Character: 11},
	}
	opts := &Options{ClientIdentifier: clients.IdentifierGeneric}

	result, err := NewPolicy(t.Context(), store).Run(t.Context(), c, params, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	test.AssertLabels(t, result, []string{"user"})
}

func TestPolicyProvider_Example2(t *testing.T) {
	t.Parallel()

	file1 := ast.MustParseModule("package example\n\nfoo := true\n")
	file2 := ast.MustParseModule("package example2\n\nimport data.example\n")

	store := inmem.NewFromObject(map[string]any{
		"workspace": map[string]any{
			"parsed": map[string]any{
				"file:///file1.rego": file1,
				"file:///file2.rego": file2,
			},
			"defined_refs": map[string][]string{
				"file:///file1.rego": {"example.foo"},
				"file:///file2.rego": {},
			},
		},
	})

	fileEdited := "package example2\n\nimport data.example\n\nallow if {\n\tfoo :=\n}\n"

	c := cache.NewCache()
	c.SetFileContents("file:///file2.rego", fileEdited)

	params := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{URI: "file:///file2.rego"},
		Position:     types.Position{Line: 5, Character: 11},
	}
	opts := &Options{ClientIdentifier: clients.IdentifierGeneric}

	result, err := NewPolicy(t.Context(), store).Run(t.Context(), c, params, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	test.AssertLabels(t, result, []string{"input", "example", "example.foo"})
}
