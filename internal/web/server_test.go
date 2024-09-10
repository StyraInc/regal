package web

import (
	"bytes"
	"strings"
	"testing"
)

func TestTemplateFoundAndParsed(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}

	st := state{Code: "package main\n\nimport rego.v1\n"}

	err := tpl.ExecuteTemplate(&buf, mainTemplate, st)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(buf.String(), "<!DOCTYPE html>") {
		t.Fatalf("expected HTML document, got %s", buf.String())
	}
}
