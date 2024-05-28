package providers

import (
	"slices"
	"strings"
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/completions/refs"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/parse"
)

func TestPackageRefs(t *testing.T) {
	t.Parallel()

	c := cache.NewCache()

	regoFiles := map[string]string{
		"file:///foo/bar/file.rego": `package foo.bar

allow := true
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

	for uri, contents := range regoFiles {
		mod := parse.MustParseModule(contents)
		c.SetModule(uri, mod)

		c.SetFileRefs(uri, refs.ForModule(mod))
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

	completions, err := p.Run(c, completionParams, nil)
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
