package parse

import "testing"

func TestParseModule(t *testing.T) {
	t.Parallel()

	parsed, err := Module("test.rego", `package p`)
	if err != nil {
		t.Fatal(err)
	}

	if exp, got := "data.p", parsed.Package.Path.String(); exp != got {
		t.Errorf("expected %q, got %q", exp, got)
	}
}
