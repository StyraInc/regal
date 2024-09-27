package lsp

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/open-policy-agent/opa/storage"

	"github.com/styrainc/regal/internal/parse"
)

func TestPutFileModStoresRoastRepresentation(t *testing.T) {
	t.Parallel()

	store := NewRegalStore()
	ctx := context.Background()
	fileURI := "file:///example.rego"
	module := parse.MustParseModule("package example\n\nrule := true")

	if err := PutFileMod(ctx, store, fileURI, module); err != nil {
		t.Fatalf("PutFileMod failed: %v", err)
	}

	parsed, err := storage.ReadOne(ctx, store, storage.Path{"workspace", "parsed", fileURI})
	if err != nil {
		t.Fatalf("store.Read failed: %v", err)
	}

	pretty, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent failed: %v", err)
	}

	// This is certainly testing the implementation rather than the behavior, but we actually
	// want some tests to fail if the implementation changes, so we don't have to chase this
	// down elsewhere.
	expect := `{
  "package": {
    "location": "1:1:cGFja2FnZQ==",
    "path": [
      {
        "location": "1:9:ZXhhbXBsZQ==",
        "type": "var",
        "value": "data"
      },
      {
        "location": "1:9:ZXhhbXBsZQ==",
        "type": "string",
        "value": "example"
      }
    ]
  },
  "rules": [
    {
      "head": {
        "assign": true,
        "location": "3:1:cnVsZSA6PSB0cnVl",
        "name": "rule",
        "ref": [
          {
            "location": "3:1:cnVsZQ==",
            "type": "var",
            "value": "rule"
          }
        ],
        "value": {
          "location": "3:9:dHJ1ZQ==",
          "type": "boolean",
          "value": true
        }
      },
      "location": "3:1:cnVsZSA6PSB0cnVl"
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
	ctx := context.Background()
	fileURI := "file:///example.rego"

	if err := PutFileRefs(ctx, store, fileURI, []string{"foo", "bar"}); err != nil {
		t.Fatalf("PutFileRefs failed: %v", err)
	}

	value, err := storage.ReadOne(ctx, store, storage.Path{"workspace", "defined_refs", fileURI})
	if err != nil {
		t.Fatalf("store.Read failed: %v", err)
	}

	arr, ok := value.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", value)
	}

	if !slices.Equal(arr, []string{"foo", "bar"}) {
		t.Fatalf("expected [foo bar], got %v", arr)
	}
}
