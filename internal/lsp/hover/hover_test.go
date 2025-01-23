package hover

import (
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/types"

	"github.com/styrainc/regal/internal/testutil"
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
		{
			&ast.Builtin{
				Name:        "foo.bar",
				Description: "Description for Foo Bar",
				Decl: types.NewFunction(
					types.Args(
						types.Named("arg1", types.S).Description("arg1 for foobar"),
						types.Named("arg2", types.S).Description("arg2 for foobar"),
					),
					types.Named("output", types.N).Description("the output for foobar"),
				),
				Categories: []string{"foo", "url=https://example.com"},
			},
			"testdata/hover/foobar.md",
		},
	}

	for _, c := range cases {
		file := testutil.MustReadFile(t, c.testdata)
		hoverContent := CreateHoverContent(c.builtin)

		if string(file) != hoverContent {
			t.Errorf("Expected %s, got %s", string(file), hoverContent)
		}
	}
}
