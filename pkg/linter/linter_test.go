package linter

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/test"
)

func TestLintBasic(t *testing.T) {
	t.Parallel()

	policy := `package p

# TODO: fix this
camelCase {
	1 == 1
}
`

	linter := NewLinter().WithAddedBundle(test.GetRegalBundle(t))

	result, err := linter.Lint(context.Background(), test.InputPolicy("p.rego", policy))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "todo-comment" {
		t.Errorf("excpected first violation to be 'todo-comments', got %s", result.Violations[0].Title)
	}

	if result.Violations[0].Location.Row != 3 {
		t.Errorf("excpected first violation to be on line 3, got %d", result.Violations[0].Location.Row)
	}

	if result.Violations[0].Location.Column != 1 {
		t.Errorf("excpected first violation to be on column 1, got %d", result.Violations[0].Location.Column)
	}

	if string(result.Violations[0].Location.Text) != "# TODO: fix this" {
		t.Errorf("excpected first violation to be on '# TODO: fix this', got %s", result.Violations[0].Location.Text)
	}

	if result.Violations[1].Title != "prefer-snake-case" {
		t.Errorf("excpected second violation to be 'prefer-snake-case', got %s", result.Violations[1].Title)
	}

	if result.Violations[1].Location.Row != 4 {
		t.Errorf("excpected second violation to be on line 4, got %d", result.Violations[1].Location.Row)
	}

	if result.Violations[1].Location.Column != 1 {
		t.Errorf("excpected second violation to be on column 1, got %d", result.Violations[1].Location.Column)
	}

	if string(result.Violations[1].Location.Text) != "camelCase {" {
		t.Errorf("excpected second violation to be on 'camelCase {', got %s", result.Violations[1].Location.Text)
	}
}

func TestLintWithUserConfig(t *testing.T) {
	t.Parallel()

	policy := `package p

# TODO: fix this
camelCase {
	1 == 1
}
`

	configRaw := `rules:
  comments:
    todo-comment:
      enabled: false`

	config := rio.MustYAMLToMap(strings.NewReader(configRaw))

	linter := NewLinter().
		WithUserConfig(config).
		WithAddedBundle(test.GetRegalBundle(t))

	result, err := linter.Lint(context.Background(), test.InputPolicy("p.rego", policy))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "prefer-snake-case" {
		t.Errorf("excpected first violation to be 'prefer-snake-case', got %s", result.Violations[0].Title)
	}
}

func TestLintWithCustomRule(t *testing.T) {
	t.Parallel()

	policy := "package p\n"

	linter := NewLinter().
		WithAddedBundle(test.GetRegalBundle(t)).
		WithCustomRules([]string{filepath.Join("testdata", "custom.rego")})

	result, err := linter.Lint(context.Background(), test.InputPolicy("p.rego", policy))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Violations) != 1 {
		t.Errorf("expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "acme-corp-package" {
		t.Errorf("excpected first violation to be 'acme-corp-package', got %s", result.Violations[0].Title)
	}
}
