package encoding

import (
	"testing"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

var pkg = &ast.Package{
	Location: &ast.Location{
		Row:  6,
		Col:  1,
		Text: []byte("foo"),
	},
	Path: ast.Ref{
		ast.DefaultRootDocument,
		ast.StringTerm("foo"),
	},
}

func TestAnnotationsOnPackage(t *testing.T) {
	t.Parallel()

	module := ast.Module{
		Package: pkg,
		Annotations: []*ast.Annotations{
			{
				Location: &ast.Location{
					Row: 1,
					Col: 1,
				},
				Scope: "package",
				Title: "foo",
			},
		},
	}

	json := jsoniter.ConfigFastest

	roast, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal annotations: %v", err)
	}

	// package annotations should end up on the package object
	// and *not* on the module object, contrary to how OPA
	// currently does it

	expected := `{
  "package": {
    "location": "6:1:6:4",
    "path": [
      {
        "type": "var",
        "value": "data"
      },
      {
        "type": "string",
        "value": "foo"
      }
    ],
    "annotations": [
      {
        "location": "1:1:1:1",
        "scope": "package",
        "title": "foo"
      }
    ]
  }
}`

	if string(roast) != expected {
		t.Fatalf("expected %s but got %s", expected, roast)
	}
}

func TestAnnotationsOnPackageBothPackageAndSubpackagesScope(t *testing.T) {
	t.Parallel()

	module := ast.Module{
		Package: pkg,
		Annotations: []*ast.Annotations{
			{
				Location: &ast.Location{
					Row: 1,
					Col: 1,
				},
				Scope: "package",
				Title: "foo",
			},
			{
				Location: &ast.Location{
					Row: 3,
					Col: 1,
				},
				Scope: "subpackages",
				Title: "bar",
			},
		},
	}

	json := jsoniter.ConfigFastest

	roast, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal annotations: %v", err)
	}

	expected := `{
  "package": {
    "location": "6:1:6:4",
    "path": [
      {
        "type": "var",
        "value": "data"
      },
      {
        "type": "string",
        "value": "foo"
      }
    ],
    "annotations": [
      {
        "location": "1:1:1:1",
        "scope": "package",
        "title": "foo"
      },
      {
        "location": "3:1:3:1",
        "scope": "subpackages",
        "title": "bar"
      }
    ]
  }
}`

	if string(roast) != expected {
		t.Fatalf("expected %s but got %s", expected, roast)
	}
}

func TestRuleAndDocumentScopedAnnotationsOnPackageAreDropped(t *testing.T) {
	t.Parallel()

	module := ast.Module{
		Package: pkg,
		Annotations: []*ast.Annotations{
			{
				Location: &ast.Location{
					Row: 1,
					Col: 1,
				},
				Scope: "package",
				Title: "foo",
			},
			{
				Location: &ast.Location{
					Row: 3,
					Col: 1,
				},
				Scope: "rule",
				Title: "bar",
			},
			{
				Location: &ast.Location{
					Row: 4,
					Col: 1,
				},
				Scope: "document",
				Title: "baz",
			},
		},
	}

	json := jsoniter.ConfigFastest

	roast, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal annotations: %v", err)
	}

	expected := `{
  "package": {
    "location": "6:1:6:4",
    "path": [
      {
        "type": "var",
        "value": "data"
      },
      {
        "type": "string",
        "value": "foo"
      }
    ],
    "annotations": [
      {
        "location": "1:1:1:1",
        "scope": "package",
        "title": "foo"
      }
    ]
  }
}`

	if string(roast) != expected {
		t.Fatalf("expected %s but got %s", expected, roast)
	}
}

func TestSerializedModuleSize(t *testing.T) {
	t.Parallel()

	policy := mustReadTestFile(t, "testdata/policy.rego")
	module := ast.MustParseModuleWithOpts(string(policy), ast.ParserOptions{
		ProcessAnnotation: true,
	})

	json := jsoniter.ConfigFastest

	roast, err := json.Marshal(module)
	if err != nil {
		t.Fatalf("failed to marshal module: %v", err)
	}

	// This test will fail whenever the size of the serialized module changes,
	// which not often and when it happens it's good to know about it, update
	// and move on.
	if len(roast) != 85979 {
		t.Fatalf("expected %d but got %d", 85979, len(roast))
	}
}

// BenchmarkSerializeModule-10    	    2281	    500175 ns/op	  219349 B/op	    9883 allocs/op
// BenchmarkSerializeModule-10    	    2488	    479095 ns/op	  217090 B/op	    9805 allocs/op

func BenchmarkSerializeModule(b *testing.B) {
	policy := mustReadTestFile(b, "testdata/policy.rego")
	module := ast.MustParseModuleWithOpts(string(policy), ast.ParserOptions{
		ProcessAnnotation: true,
	})

	json := jsoniter.ConfigFastest

	b.ResetTimer()

	for range b.N {
		_, err := json.Marshal(module)
		if err != nil {
			b.Fatalf("failed to marshal module: %v", err)
		}
	}
}
