package hints

import (
	"slices"
	"testing"

	"github.com/open-policy-agent/regal/internal/parse"
)

func TestHints(t *testing.T) {
	t.Parallel()

	mod := `package foo

incomplete`

	_, err := parse.Module("test.rego", mod)
	if err == nil {
		t.Fatal("expected error")
	}

	hints, err := GetForError(err)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expectedHints := []string{"rego-parse-error/var-cannot-be-used-for-rule-name"}
	if !slices.Equal(hints, expectedHints) {
		t.Fatalf("expected\n%v but got\n%v", expectedHints, hints)
	}
}
