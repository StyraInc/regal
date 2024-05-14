package lsp

import (
	"testing"

	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/lsp/types/symbols"
)

func toStrPtr(s string) *string {
	return &s
}

func TestRefToString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		title    string
		ref      ast.Ref
		expected string
	}{
		{
			"single var",
			ast.Ref{
				ast.VarTerm("foo"),
			},
			"foo",
		},
		{
			"var in middle",
			ast.Ref{
				ast.StringTerm("foo"),
				ast.VarTerm("bar"),
				ast.StringTerm("baz"),
			},
			"foo[bar].baz",
		},
		{
			"strings",
			ast.Ref{
				ast.DefaultRootDocument,
				ast.StringTerm("foo"),
				ast.StringTerm("bar"),
				ast.StringTerm("baz"),
			},
			"data.foo.bar.baz",
		},
		{
			"consecutive vars",
			ast.Ref{
				ast.VarTerm("foo"),
				ast.VarTerm("bar"),
				ast.VarTerm("baz"),
			},
			"foo[bar][baz]",
		},
		{
			"mixed",
			ast.Ref{
				ast.VarTerm("foo"),
				ast.VarTerm("bar"),
				ast.StringTerm("baz"),
				ast.VarTerm("qux"),
				ast.StringTerm("quux"),
			},
			"foo[bar].baz[qux].quux",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			result := refToString(tc.ref)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestDocumentSymbols(t *testing.T) {
	t.Parallel()

	cases := []struct {
		title    string
		policy   string
		expected types.DocumentSymbol
	}{
		{
			"only package",
			`package foo`,
			types.DocumentSymbol{
				Name: "data.foo",
				Kind: symbols.Package,
				Range: types.Range{
					Start: types.Position{Line: 0, Character: 0},
					End:   types.Position{Line: 0, Character: 11},
				},
			},
		},
		{
			"call",
			`package p

			i := indexof("a", "a")`,
			types.DocumentSymbol{
				Name: "data.p",
				Kind: symbols.Package,
				Range: types.Range{
					Start: types.Position{Line: 0, Character: 0},
					End:   types.Position{Line: 2, Character: 25},
				},
				Children: &[]types.DocumentSymbol{
					{
						Name:   "i",
						Kind:   symbols.Variable,
						Detail: toStrPtr("single-value rule (number)"),
						Range: types.Range{
							Start: types.Position{Line: 2, Character: 3},
							End:   types.Position{Line: 2, Character: 22},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			module, err := ast.ParseModule("test.rego", tc.policy)
			if err != nil {
				t.Fatal(err)
			}

			syms := documentSymbols(tc.policy, module)

			pkg := syms[0]
			if pkg.Name != tc.expected.Name {
				t.Errorf("Expected %s, got %s", tc.expected.Name, pkg.Name)
			}

			if pkg.Kind != tc.expected.Kind {
				t.Errorf("Expected %d, got %d", tc.expected.Kind, pkg.Kind)
			}

			if pkg.Range != tc.expected.Range {
				t.Errorf("Expected %v, got %v", tc.expected.Range, pkg.Range)
			}

			if pkg.Detail != tc.expected.Detail {
				t.Errorf("Expected %v, got %v", tc.expected.Detail, pkg.Detail)
			}

			if pkg.Children != nil {
				if tc.expected.Children == nil {
					t.Fatalf("Expected no children, got %v", pkg.Children)
				}

				for i, child := range *pkg.Children {
					expectedChild := (*tc.expected.Children)[i]

					if child.Name != expectedChild.Name {
						t.Errorf("Expected %s, got %s", child.Name, expectedChild.Name)
					}

					if child.Kind != expectedChild.Kind {
						t.Errorf("Expected %d, got %d", expectedChild.Kind, child.Kind)
					}

					if child.Range != expectedChild.Range {
						t.Errorf("Expected %v, got %v", expectedChild.Range, child.Range)
					}

					if child.Detail != expectedChild.Detail {
						if expectedChild.Detail == nil && child.Detail != nil {
							t.Errorf("Expected detail to be nilgot %v", child.Detail)
						} else if *child.Detail != *expectedChild.Detail {
							t.Errorf("Expected %s, got %s", *expectedChild.Detail, *child.Detail)
						}
					}
				}
			}
		})
	}
}

func TestSimplifyType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    string
		expected string
	}{
		{
			"set",
			"set",
		},
		{
			"set[any]",
			"set",
		},
		{
			"any<set, object>",
			"any",
		},
		{
			"output: any<set[any], object>",
			"any",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			result := simplifyType(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}
