package providers

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/completions/refs"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/parse"
)

func TestUsedRefs(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	currentlyEditingFileContents := `package example

import rego.v1

import data.foo as wow
import data.bar

allow if input.user == "admin"

allow if data.users.admin == input.user

deny contains wow.password if {
	input.magic == true
}

deny contains input.parrot if {
	bar.parrot != "a bird"
}
`

	uri := "file:///example.rego"

	mod, err := parse.Module(uri, currentlyEditingFileContents)
	if err != nil {
		t.Fatalf("Unexpected error when parsing %s contents: %v", uri, err)
	}

	c.SetModule(uri, mod)
	c.SetFileRefs(uri, refs.DefinedInModule(mod))

	refNames, err := refs.UsedInModule(context.Background(), mod)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	c.SetUsedRefs(uri, refNames)

	c.SetFileContents(uri, currentlyEditingFileContents+`
allow if {
  true == i
}`)

	p := &UsedRefs{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: uri,
		},
		Position: types.Position{
			Line:      20,
			Character: 11,
		},
	}

	completions, err := p.Run(c, completionParams, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedRefs := []string{
		"input.magic",
		"input.parrot",
		"input.user",
	}
	slices.Sort(expectedRefs)

	foundRefs := make([]string, len(completions))

	for i, c := range completions {
		foundRefs[i] = c.Label
	}

	slices.Sort(foundRefs)

	if !slices.Equal(expectedRefs, foundRefs) {
		t.Fatalf(
			"Expected completions to be\n%s\ngot:\n%s",
			strings.Join(expectedRefs, "\n"),
			strings.Join(foundRefs, "\n"),
		)
	}
}
