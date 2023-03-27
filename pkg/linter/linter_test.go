package linter

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/test"
)

func TestLintBasic(t *testing.T) {
	t.Parallel()

	policy := parse.MustParseModule(`package p
	# TODO: fix this
	camelCase {
		true
	}`)

	linter := NewLinter().WithAddedBundle(test.GetRegalBundle(t))

	result, err := linter.Lint(context.Background(), map[string]*ast.Module{
		"p.rego": policy,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(result.Violations))
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

	policy := parse.MustParseModule(`package p
	# TODO: fix this
	camelCase {
		true
	}`)

	configRaw := `rules:
  comments:
    todo-comment:
      enabled: false`

	config := rio.MustYAMLToMap(strings.NewReader(configRaw))

	linter := NewLinter().
		WithUserConfig(config).
		WithAddedBundle(test.GetRegalBundle(t))

	result, err := linter.Lint(context.Background(), map[string]*ast.Module{
		"p.rego": policy,
	})
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

	policy := parse.MustParseModule(`package p`)

	linter := NewLinter().
		WithAddedBundle(test.GetRegalBundle(t)).
		WithCustomRules([]string{filepath.Join("testdata", "custom.rego")})

	result, err := linter.Lint(context.Background(), map[string]*ast.Module{
		"p.rego": policy,
	})
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
