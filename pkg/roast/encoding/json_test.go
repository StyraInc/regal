package encoding

import (
	"bytes"
	"encoding/json/jsontext"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/open-policy-agent/regal/internal/encoding/noast"
	"github.com/open-policy-agent/regal/internal/test/must"
)

// Simple routine check to see that things are working as expected.
// While it would be good to add tests for the encoding of each AST node individually, this is thoroughly tested
// via Regal as it consumes the Roast format extensively.

func TestJsonLocationEncoding(t *testing.T) {
	t.Parallel()

	must.Marshal(t, ast.MustParseModuleWithOpts(`
package p

import rego.v1

import data.foo.bar

# METADATA
# description: foo bar went to the bar
allow if true

# regular comment

add(x, y) := x + y

partial[x] contains y if {
	some x, y in input

	every z in x {
		z == y
	}
}

obj := {"foo": {"number": 1}, "string": {"set"}, "bool": false}

arr := [1, {"foo": {"key": 1}}]

sc := {x | x := [1, 2, 3][_]}

ac := [x | x := [1, 2, 3][_]]

oc := {k:v | some k, v in input}

test_foo if {
	allow with input as {"foo": "bar"}
}
	`, ast.ParserOptions{ProcessAnnotation: true}))
}

// https://github.com/open-policy-agent/regal/issues/1592
func TestJSONRoundTripBigNumber(t *testing.T) {
	t.Parallel()

	mod := ast.MustParseModule("package p\n\nn := 1e400")
	val := must.Return(JSONRoundTripTo[ast.Value](mod, noast.Options))(t)
	obj := must.Be[ast.Object](t, val)
	num := must.Be[ast.Number](t, must.Return(obj.Find(ast.Ref{
		ast.InternedTerm("rules"),
		ast.InternedTerm(0),
		ast.InternedTerm("head"),
		ast.InternedTerm("value"),
		ast.InternedTerm("value"),
	}))(t))

	if !num.Equal(ast.Number("1e400")) {
		t.Fatalf("expected n to be 1e400, got: %v", num)
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	// Example test cases for the Get function
	tests := []struct {
		name string
		json string
		path jsontext.Pointer
		want jsontext.Value
	}{
		{
			name: "object key",
			json: `{"foo": "bar"}`,
			path: "/foo",
			want: jsontext.Value(`"bar"`),
		},
		{
			name: "nested object key",
			json: `{"foo": {"bar": "baz"}}`,
			path: "/foo/bar",
			want: jsontext.Value(`"baz"`),
		},
		{
			name: "nested object second object",
			json: `{"position":{"character":16,"line":5},"textDocument":{"uri":"file:///foo.txt"}}`,
			path: "/textDocument/uri",
			want: jsontext.Value(`"file:///foo.txt"`),
		},
		{
			name: "non-existent key",
			json: `{"foo": "bar"}`,
			path: "/baz",
			want: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := Get(jsontext.Value(test.json), test.path)
			if !bytes.Equal(got, test.want) {
				t.Fatalf("expected:\n%v\n got:\n%v", test.want, got)
			}
		})
	}
}
