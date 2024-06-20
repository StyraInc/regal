package providers

import (
	"slices"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/storage/inmem"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/clients"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/parse"
)

func TestPolicyProvider_Example1(t *testing.T) {
	t.Parallel()

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

	store := inmem.NewFromObject(map[string]interface{}{
		"workspace": map[string]interface{}{
			"parsed": map[string]interface{}{
				testCaseFileURI: module,
			},
		},
	})

	locals := NewPolicy(store)

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

	labels := []string{}
	for _, item := range result {
		labels = append(labels, item.Label)
	}

	expected := []string{"user"}
	if !slices.Equal(expected, labels) {
		t.Fatalf("expected %v, got %v", expected, labels)
	}
}

func TestPolicyProvider_Example2(t *testing.T) {
	t.Parallel()

	file1 := ast.MustParseModule(`package example

import rego.v1

foo := true
`)
	file2 := ast.MustParseModule(`package example2

import rego.v1

import data.example
`)

	store := inmem.NewFromObject(map[string]interface{}{
		"workspace": map[string]interface{}{
			"parsed": map[string]interface{}{
				"file:///file1.rego": file1,
				"file:///file2.rego": file2,
			},
			"defined_refs": map[string][]string{
				"file:///file1.rego": {"example.foo"},
				"file:///file2.rego": {},
			},
		},
	})

	locals := NewPolicy(store)
	fileEdited := `package example2

import rego.v1

import data.example

allow if {
	foo :=
}
`
	c := cache.NewCache()

	c.SetFileContents("file:///file2.rego", fileEdited)

	params := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: "file:///file2.rego",
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

	labels := []string{}
	for _, item := range result {
		labels = append(labels, item.Label)
	}

	expected := []string{"example", "example.foo"}
	if !slices.Equal(expected, labels) {
		t.Fatalf("expected %v, got %v", expected, labels)
	}
}
