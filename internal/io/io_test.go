package io

import (
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
