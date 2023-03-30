package io

import (
	"strings"
	"testing"
)

func TestJSONRoundTrip(t *testing.T) {
	t.Parallel()

	type foo struct {
		Bar string `json:"bar"`
	}

	m := map[string]any{"bar": "foo"}
	f := foo{}

	if err := JSONRoundTrip(m, &f); err != nil {
		t.Fatal(err)
	}

	if f.Bar != "foo" {
		t.Errorf("expected JSON roundtrip to set struct value")
	}
}

func TestReadRow(t *testing.T) {
	t.Parallel()

	text := `a b c
d e f
g h i
j k l`

	bs, err := ReadRow(strings.NewReader(text), 3)
	if err != nil {
		t.Fatal(err)
	}

	if string(bs) != "g h i" {
		t.Errorf("expected row 3 to be 'g h i'")
	}
}
