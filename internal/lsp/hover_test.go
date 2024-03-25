package lsp

import (
	"os"
	"testing"

	"github.com/open-policy-agent/opa/ast"
)

func TestCreateHoverContent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		builtin  *ast.Builtin
		testdata string
	}{
		{
			ast.IndexOf,
			"testdata/hover/indexof.md",
		},
		{
			ast.ReachableBuiltin,
			"testdata/hover/graphreachable.md",
		},
		{
			ast.JSONFilter,
			"testdata/hover/jsonfilter.md",
		},
	}

	for _, c := range cases {
		file, err := os.ReadFile(c.testdata)
		if err != nil {
			t.Fatal(err)
		}

		hoverContent := createHoverContent(c.builtin)

		if string(file) != hoverContent {
			t.Errorf("Expected %s, got %s", string(file), hoverContent)
		}
	}
}
