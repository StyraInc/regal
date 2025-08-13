package lsp

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/storage"

	"github.com/styrainc/regal/internal/parse"
)

type illegalResolver struct{}

func (illegalResolver) Resolve(ref ast.Ref) (any, error) {
	return nil, fmt.Errorf("illegal value: %v", ref)
}

func TestPutFileModStoresRoastRepresentation(t *testing.T) {
	t.Parallel()

	store := NewRegalStore()
	fileURI := "file:///example.rego"
	module := parse.MustParseModule("package example\n\nrule := true")

	if err := PutFileMod(t.Context(), store, fileURI, module); err != nil {
		t.Fatalf("PutFileMod failed: %v", err)
	}

	parsed, err := storage.ReadOne(t.Context(), store, storage.Path{"workspace", "parsed", fileURI})
	if err != nil {
		t.Fatalf("store.Read failed: %v", err)
	}

	parsedVal, ok := parsed.(ast.Value)
	if !ok {
		t.Fatalf("expected ast.Value, got %T", parsed)
	}

	parsedMap, err := ast.ValueToInterface(parsedVal, illegalResolver{})
	if err != nil {
		t.Fatalf("ast.ValueToInterface failed: %v", err)
	}

	pretty, err := json.MarshalIndent(parsedMap, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent failed: %v", err)
	}

	// This is certainly testing the implementation rather than the behavior, but we actually
	// want some tests to fail if the implementation changes, so we don't have to chase this
	// down elsewhere.
	expect := `{
  "package": {
    "location": "1:1:1:8",
    "path": [
      {
        "type": "var",
        "value": "data"
      },
      {
        "location": "1:9:1:16",
        "type": "string",
        "value": "example"
      }
    ]
  },
  "rules": [
    {
      "head": {
        "assign": true,
        "location": "3:1:3:13",
        "ref": [
          {
            "location": "3:1:3:5",
            "type": "var",
            "value": "rule"
          }
        ],
        "value": {
          "location": "3:9:3:13",
          "type": "boolean",
          "value": true
        }
      },
      "location": "3:1:3:13"
    }
  ]
}`

	if string(pretty) != expect {
		t.Errorf("expected %s, got %s", expect, pretty)
	}
}

func TestPutFileRefs(t *testing.T) {
	t.Parallel()

	store := NewRegalStore()
	fileURI := "file:///example.rego"

	if err := PutFileRefs(t.Context(), store, fileURI, []string{"foo", "bar"}); err != nil {
		t.Fatalf("PutFileRefs failed: %v", err)
	}

	value, err := storage.ReadOne(t.Context(), store, storage.Path{"workspace", "defined_refs", fileURI})
	if err != nil {
		t.Fatalf("store.Read failed: %v", err)
	}

	arr, ok := value.(*ast.Array)
	if !ok {
		t.Fatalf("expected *ast.Array, got %T", value)
	}

	expected := ast.NewArray(ast.StringTerm("foo"), ast.StringTerm("bar"))
	if !expected.Equal(arr) {
		t.Errorf("expected %v, got %v", expected, arr)
	}
}

func TestPutBuiltins(t *testing.T) {
	t.Parallel()

	store := NewRegalStore()
	builtins := map[string]*ast.Builtin{"count": ast.Count}

	if err := PutBuiltins(t.Context(), store, builtins); err != nil {
		t.Fatalf("PutBuiltins failed: %v", err)
	}

	value, err := storage.ReadOne(t.Context(), store, storage.Path{"workspace", "builtins", "count"})
	if err != nil {
		t.Fatalf("store.Read failed: %v", err)
	}

	if value == nil {
		t.Errorf("expected count builtin to exist in store")
	}
}
