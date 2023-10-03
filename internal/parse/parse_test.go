package parse

import (
	"testing"

	"github.com/styrainc/regal/internal/testutil"
)

func TestParseModule(t *testing.T) {
	t.Parallel()

	parsed := testutil.Must(Module("test.rego", `package p`))(t)

	if exp, got := "data.p", parsed.Package.Path.String(); exp != got {
		t.Errorf("expected %q, got %q", exp, got)
	}
}
