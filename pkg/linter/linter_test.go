package linter

import (
	"context"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/test"
)

func TestLintBasic(t *testing.T) {
	t.Parallel()

	policy := `package p
	# TODO: fix this
	camelCase {
		true
	}`

	linter := NewLinter().WithAddedBundle(test.GetRegalBundle(t))

	result, err := linter.Lint(context.Background(), LoaderResultFromString(policy))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Violations) != 2 {
		t.Errorf("expected 2 violations, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "todo-comment" {
		t.Errorf("excpected first violation to be 'todo-comments', got %s", result.Violations[0].Title)
	}

	if result.Violations[1].Title != "prefer-snake-case" {
		t.Errorf("excpected second violation to be 'prefer-snake-case', got %s", result.Violations[1].Title)
	}
}

func TestLintWithUserConfig(t *testing.T) {
	t.Parallel()

	policy := `package p
	# TODO: fix this
	camelCase {
		true
	}`

	configRaw := `rules:
  comments:
    todo-comment:
      enabled: false`

	config := rio.MustYAMLToMap(strings.NewReader(configRaw))

	userBundle := LoaderResultFromString(policy)

	linter := NewLinter().
		WithUserConfig(config).
		WithAddedBundle(test.GetRegalBundle(t))

	result, err := linter.Lint(context.Background(), userBundle)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Violations) != 1 {
		t.Errorf("expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "prefer-snake-case" {
		t.Errorf("excpected first violation to be 'prefer-snake-case', got %s", result.Violations[0].Title)
	}
}

func LoaderResultFromString(policy string) *loader.Result {
	return &loader.Result{Modules: map[string]*loader.RegoFile{
		"p.rego": {
			Name:   "p.rego",
			Parsed: ast.MustParseModule(policy),
			Raw:    []byte(policy),
		},
	}}
}
