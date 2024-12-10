package providers

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/completions/refs"
	"github.com/styrainc/regal/internal/lsp/rego"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/parse"
)

func TestPackageRefs(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	regoFiles := map[string]string{
		"file:///foo/bar/file.rego": `package foo.bar

`,
		"file:///foo/baz/file.rego": `package foo.baz

bax(x) := true
`,
		// if a ref for a package is in more than one file,
		// then we must ensure it's not suggested as a completion
		// in either file
		"file:///bar/file1.rego": "package bar\nallow := true\n",
		"file:///bar/file2.rego": `package bar`,
	}

	// string replace to deal with editors stripping whitespace
	fileContents := strings.ReplaceAll(`package bar

import

`, "import", "import ")

	c.SetFileContents("file:///bar/file2.rego", fileContents)

	builtins := rego.BuiltinsForCapabilities(ast.CapabilitiesForThisVersion())

	for uri, contents := range regoFiles {
		mod := parse.MustParseModule(contents)
		c.SetModule(uri, mod)

		c.SetFileRefs(uri, refs.DefinedInModule(mod, builtins))
	}

	p := &PackageRefs{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: "file:///bar/file2.rego",
		},
		Position: types.Position{
			Line:      2,
			Character: 8,
		},
	}

	completions, err := p.Run(context.Background(), c, completionParams, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// data.foo is not defined in a file, but it included as a partial ref
	// as it's still valid for import and might be helpful.
	expectedRefs := []string{"data.foo", "data.foo.bar", "data.foo.baz"}
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

func TestPackageRefs_Metadata(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	regoFiles := map[string]string{
		"file:///foo/foo.rego": `# METADATA
# scope: subpackages
# authors:
# - Foo
# related_resources:
# - https://example.com
package foo
`,
		"file:///foo/bar/bar.rego": `# METADATA
# scope: package
# title: Bar
# authors:
# - Bar
package foo.bar
`,
	}

	// string replace to deal with editors stripping whitespace
	fileContents := strings.ReplaceAll(`package example

import

`, "import", "import ")

	c.SetFileContents("file:///file.rego", fileContents)

	builtins := rego.BuiltinsForCapabilities(ast.CapabilitiesForThisVersion())

	for uri, contents := range regoFiles {
		mod := parse.MustParseModule(contents)
		c.SetModule(uri, mod)

		c.SetFileRefs(uri, refs.DefinedInModule(mod, builtins))
	}

	p := &PackageRefs{}

	completionParams := types.CompletionParams{
		TextDocument: types.TextDocumentIdentifier{
			URI: "file:///file.rego",
		},
		Position: types.Position{
			Line:      2,
			Character: 8,
		},
	}

	completions, err := p.Run(context.Background(), c, completionParams, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedRefs := map[string]struct {
		Label            string
		Authors          []string
		RelatedResources []string
	}{
		"data.foo": {
			Authors:          []string{"Foo"},
			RelatedResources: []string{"https://example.com"},
		},
		"data.foo.bar": {
			Authors: []string{"Bar"},
		},
	}

	for _, c := range completions {
		ref, ok := expectedRefs[c.Label]
		if !ok {
			t.Fatalf("Unexpected completion: %s", c.Label)
		}

		for _, a := range ref.Authors {
			if !strings.Contains(c.Documentation.Value, "* "+a) {
				t.Fatalf("Expected author %s to be in documentation, got: %s", a, c.Documentation.Value)
			}
		}

		for _, a := range ref.RelatedResources {
			if !strings.Contains(c.Documentation.Value, fmt.Sprintf("* [%s]", a)) {
				t.Fatalf("Expected related resource %s to be in documentation, got: %s", a, c.Documentation.Value)
			}
		}
	}
}
