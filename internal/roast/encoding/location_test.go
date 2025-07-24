package encoding

import (
	"fmt"
	"testing"

	jsoniter "github.com/json-iterator/go"

	"github.com/open-policy-agent/opa/v1/ast"
)

func TestLocation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		location ast.Location
		expected string
	}{
		{
			name: "multiple lines",
			location: ast.Location{
				Row:  5,
				Col:  2,
				Text: []byte("allow if {\n	input.foo == true\n}"),
			},
			expected: "5:2:7:2",
		},
		{
			name: "single line",
			location: ast.Location{
				Row:  1,
				Col:  1,
				Text: []byte("package example"),
			},
			expected: "1:1:1:16",
		},
	}

	json := jsoniter.ConfigFastest

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stream := json.BorrowStream(nil)
			defer json.ReturnStream(stream)

			stream.WriteVal(tc.location)

			if string(stream.Buffer()) != fmt.Sprintf("\"%s\"", tc.expected) {
				t.Fatalf("expected %s but got %s", tc.expected, string(stream.Buffer()))
			}
		})
	}
}

func TestLocationHeadValue(t *testing.T) {
	// Separate test for this as we found the end position would sometimes be off,
	// e.g. the end column would be presented as before the start column.
	t.Parallel()

	module := ast.MustParseModule("package foo.bar\n\nrule := true")
	json := jsoniter.ConfigFastest

	out, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal module: %v", err)
	}

	expect := `{
  "package": {
    "location": "1:1:1:8",
    "path": [
      {
        "type": "var",
        "value": "data"
      },
      {
        "location": "1:9:1:12",
        "type": "string",
        "value": "foo"
      },
      {
        "location": "1:13:1:16",
        "type": "string",
        "value": "bar"
      }
    ]
  },
  "rules": [
    {
      "location": "3:1:3:13",
      "head": {
        "location": "3:1:3:13",
        "ref": [
          {
            "location": "3:1:3:5",
            "type": "var",
            "value": "rule"
          }
        ],
        "assign": true,
        "value": {
          "location": "3:9:3:13",
          "type": "boolean",
          "value": true
        }
      }
    }
  ]
}`
	if string(out) != expect {
		t.Fatalf("expected %s but got %s", expect, out)
	}
}
