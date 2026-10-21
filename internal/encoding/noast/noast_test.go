package noast

import (
	"encoding/json/v2"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/util"

	"github.com/open-policy-agent/regal/internal/test/must"
)

var ValueTests = []struct {
	name string
	json []byte
	want ast.Value
}{
	{
		name: "not interned string",
		json: []byte(`"not interned"`),
		want: ast.String("not interned"),
	},
	{
		name: "interned string",
		json: []byte(`"interned"`),
		want: ast.InternedValue("interned"),
	},
	{
		name: "string with escapes",
		json: []byte(`"data.regal.rules.bugs[\"argument-always-wildcard_test\"]"`),
		want: ast.InternedValue(`data.regal.rules.bugs["argument-always-wildcard_test"]`),
	},
	{
		name: "boolean true",
		json: []byte(`true`),
		want: ast.InternedBooleanTrueValue,
	},
	{
		name: "boolean false",
		json: []byte(`false`),
		want: ast.InternedBooleanFalseValue,
	},
	{
		name: "null",
		json: []byte(`null`),
		want: ast.NullValue,
	},
	{
		name: "interned integer",
		json: []byte(`50`),
		want: ast.Number("50"),
	},
	{
		name: "not interned integer",
		json: []byte(`123456789012345678901234567890`),
		want: ast.Number("123456789012345678901234567890"),
	},
	{
		name: "float",
		json: []byte(`3.14`),
		want: ast.Number("3.14"),
	},
	{
		name: "array",
		json: []byte(`[1, "two", true, null, [3, 4]]`),
		want: ast.MustParseTerm(`[1, "two", true, null, [3, 4]]`).Value,
	},
	{
		name: "empty array",
		json: []byte(`[]`),
		want: ast.InternedEmptyArrayValue,
	},
	{
		name: "object",
		json: []byte(`{"a": 1, "b": "two", "c": true, "d": null, "e": {"nested": "object"}}`),
		want: ast.MustParseTerm(
			`{"a": 1, "b": "two", "c": true, "d": null, "e": {"nested": "object"}}`).Value,
	},
	{
		name: "empty object",
		json: []byte(`{}`),
		want: ast.InternedEmptyObjectValue,
	},
	{
		name: "complex object",
		json: []byte(`{"label":"data.regal.rules.bugs[\"argument-always-wildcard_test\"]"}`),
		want: ast.MustParseTerm(`{"label":"data.regal.rules.bugs[\"argument-always-wildcard_test\"]"}`).Value,
	},
}

func TestValueNoASTRoundTrip(t *testing.T) {
	t.Parallel()

	for _, test := range ValueTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			exp := must.Marshal(t, test.want, Options)
			got := must.UnmarshalTo[ast.Value](t, exp, Options)

			if !ast.ValueEqual(test.want, got) {
				t.Fatalf("expected:\n%v\n got:\n%v", test.want, got)
			}
		})
	}
}

func TestDecodeValue(t *testing.T) {
	t.Parallel()

	regalDecodeToValue := func(bs []byte) (ast.Value, error) {
		var v ast.Value

		err := json.Unmarshal(bs, &v, Options)

		return v, err
	}

	decoders := []struct {
		name string
		fn   func([]byte) (ast.Value, error)
	}{
		{name: "regal", fn: regalDecodeToValue},
		{name: "opa", fn: opaDecodeToValue},
	}

	for _, test := range ValueTests {
		for _, decoder := range decoders {
			t.Run(test.name+" "+decoder.name, func(t *testing.T) {
				t.Parallel()

				got, err := decoder.fn(test.json)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if !ast.ValueEqual(test.want, got) {
					t.Fatalf("expected:\n%v\n got:\n%v", test.want, got)
				}
			})
		}
	}
}

func TestEncodeObjectNonStringKeys(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value ast.Value
		want  string
	}{
		"number key": {ast.MustParseTerm(`{1: "a"}`).Value, `{"1":"a"}`},
		"array key":  {ast.MustParseTerm(`{[1, 2]: "a"}`).Value, `{"[1, 2]":"a"}`},
		"bool key":   {ast.MustParseTerm(`{true: "a"}`).Value, `{"true":"a"}`},
		"null key":   {ast.MustParseTerm(`{null: "a"}`).Value, `{"null":"a"}`},
		"set key":    {ast.MustParseTerm(`{{1, 2}: "a"}`).Value, `{"{1, 2}":"a"}`},
		"object key": {ast.MustParseTerm(`{{"a": 1}: "a"}`).Value, `{"{\"a\": 1}":"a"}`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := string(must.Marshal(t, test.value, Options)); got != test.want {
				t.Fatalf("expected %s, got %s", test.want, got)
			}
		})
	}
}

func init() {
	ast.InternStringTerm("a", "b", "b", "c", "d", "e", "f", "nested", "object", "two", "interned")
}

// as found in OPA's cmd/eval.go
func opaDecodeToValue(bs []byte) (value ast.Value, err error) {
	var input any
	if err = util.Unmarshal(bs, &input); err == nil {
		value, err = ast.InterfaceToValue(input)
	}

	return value, err
}
