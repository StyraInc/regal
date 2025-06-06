package providers

import (
	"slices"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/completions/refs"
	"github.com/styrainc/regal/internal/lsp/rego"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/parse"
)

func TestRuleHead(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	regoFiles := map[string]string{
		"file:///foo/foo.rego": `package foo

default allow := false

allow if count(deny) == 0

deny contains message if {
	true
}

_internal := true

funckyfunc := true


`,
	}

	builtins := rego.BuiltinsForCapabilities(ast.CapabilitiesForThisVersion())

	for uri, contents := range regoFiles {
		mod, err := parse.Module(uri, contents)
		if err != nil {
			t.Fatalf("Unexpected error when parsing %s contents: %v", uri, err)
		}

		c.SetFileContents(uri, contents)
		c.SetModule(uri, mod)
		c.SetFileRefs(uri, refs.DefinedInModule(mod, builtins))
	}

	p := &RuleHead{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: "file:///foo/foo.rego",
		},
		Position: types.Position{
			Line:      16,
			Character: 0,
		},
	}

	completions, err := p.Run(t.Context(), c, completionParams, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedRefs := []string{"allow", "deny", "_internal", "funckyfunc"}
	slices.Sort(expectedRefs)

	foundRefs := make([]string, len(completions))

	for i, c := range completions {
		foundRefs[i] = c.Label
	}

	slices.Sort(foundRefs)

	if !slices.Equal(expectedRefs, foundRefs) {
		t.Fatalf("Expected completions to be %v, got: %v", expectedRefs, foundRefs)
	}
}
