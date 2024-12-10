package ast

import (
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
)

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
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			result := RefToString(tc.ref)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}
