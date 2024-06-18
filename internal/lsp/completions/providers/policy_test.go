package providers

import (
	"testing"

	"github.com/open-policy-agent/opa/storage/inmem"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/parse"
)

func TestPolicyProvider(t *testing.T) {
	t.Parallel()

	store := inmem.NewFromObject(map[string]interface{}{
		"workspace": map[string]interface{}{
			"parsed": map[string]interface{}{},
		},
	})

	locals := NewPolicy(store)
	policy := `package p

import rego.v1

allow if {
	user := data.users[0]
	# try completion on next line
	roles := u
}
`
	module := parse.MustParseModule(policy)
	c := cache.NewCache()

	c.SetFileContents(testCaseFileURI, policy)
	c.SetModule(testCaseFileURI, module)

	params := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: testCaseFileURI,
		},
		Position: types.Position{
			Line:      7,
			Character: 11,
		},
	}

	result, err := locals.Run(
		c,
		params,
		&Options{ClientIdentifier: clients.IdentifierGeneric},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) == 0 {
		t.Fatalf("expected completion items, got none")
	}
}
