package refs

import (
	"context"
	"slices"
	"testing"

	rparse "github.com/styrainc/regal/internal/parse"
)

func TestUsedInModule(t *testing.T) {
	t.Parallel()

	mod := rparse.MustParseModule(`
package example

import data.foo as wow
import data.bar

allow if input.user == "admin"

allow if data.users.admin == input.user

deny contains wow.password if {
	input.magic == true
}

deny contains input.parrot if {
	bar.parrot != "a bird"
}
`)

	items, err := UsedInModule(context.Background(), mod)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	expectedItems := []string{
		"wow",
		"bar",
		"bar.parrot",
		"data.users.admin",
		"input.magic",
		"input.parrot",
		"input.user",
		"wow.password",
	}

	for _, item := range expectedItems {
		if !slices.Contains(items, item) {
			t.Errorf("Expected item %q not found in items", item)
		}
	}

	for _, item := range items {
		if !slices.Contains(expectedItems, item) {
			t.Errorf("Unexpected item %q found in items", item)
		}
	}
}
