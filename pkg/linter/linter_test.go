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
	1 == input.one
}
`

	linter := NewLinter().WithAddedBundle(test.GetRegalBundle(t)).WithEnableAll(true)

	result, err := linter.Lint(context.Background(), test.InputPolicy("p.rego", policy))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "todo-comment" {
		t.Errorf("expected first violation to be 'todo-comments', got %s", result.Violations[0].Title)
	}

	if result.Violations[0].Location.Row != 3 {
		t.Errorf("expected first violation to be on line 3, got %d", result.Violations[0].Location.Row)
	}

	if result.Violations[0].Location.Column != 1 {
		t.Errorf("expected first violation to be on column 1, got %d", result.Violations[0].Location.Column)
	}

	if *result.Violations[0].Location.Text != "# TODO: fix this" {
		t.Errorf("expected first violation to be on '# TODO: fix this', got %s", *result.Violations[0].Location.Text)
	}

	if result.Violations[1].Title != "prefer-snake-case" {
		t.Errorf("expected second violation to be 'prefer-snake-case', got %s", result.Violations[1].Title)
	}

	if result.Violations[1].Location.Row != 4 {
		t.Errorf("expected second violation to be on line 4, got %d", result.Violations[1].Location.Row)
	}

	if result.Violations[1].Location.Column != 1 {
		t.Errorf("expected second violation to be on column 1, got %d", result.Violations[1].Location.Column)
	}

	if *result.Violations[1].Location.Text != "camelCase {" {
		t.Errorf("expected second violation to be on 'camelCase {', got %s", *result.Violations[1].Location.Text)
	}
}

func TestLintWithUserConfig(t *testing.T) {
	t.Parallel()

	policy := `package p

foo := input.bar[_]

or := 1
`

	configRaw := `rules:
  bugs:
    rule-shadows-builtin:
      level: ignore`

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

	if result.Violations[0].Title != "top-level-iteration" {
		t.Errorf("expected first violation to be 'top-level-iteration', got %s", result.Violations[0].Title)
	}
}

func TestLintWithUserConfigTable(t *testing.T) {
	t.Parallel()

	policy := `package p

foo := input.bar[_]

 opa_fmt := "fail"

or := 1
`
	tests := []struct {
		name            string
		configRaw       string
		filename        string
		expViolations   []string
		ignoreFilesFlag []string
	}{
		{
			name:          "baseline",
			filename:      "p.rego",
			expViolations: []string{"opa-fmt", "top-level-iteration", "rule-shadows-builtin"},
		},
		{
			name: "ignore rule",
			configRaw: `rules:
  bugs:
    rule-shadows-builtin:
      level: ignore
  style:
    opa-fmt:
      level: ignore`,
			filename:      "p.rego",
			expViolations: []string{"top-level-iteration"},
		},
		{
			name: "rule level ignore files",
			configRaw: `rules:
  bugs:
    rule-shadows-builtin:
      level: error
      ignore:
        files:
          - p.rego
  style:
    opa-fmt:
      level: error
      ignore:
        files:
          - p.rego`,
			filename:      "p.rego",
			expViolations: []string{"top-level-iteration"},
		},
		{
			name: "user config global ignore files",
			configRaw: `ignore:
  files:
    - p.rego`,
			filename:      "p.rego",
			expViolations: []string{},
		},
		{
			name:            "CLI flag ignore files",
			filename:        "p.rego",
			expViolations:   []string{},
			ignoreFilesFlag: []string{"p.rego"},
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			linter := NewLinter().
				WithAddedBundle(test.GetRegalBundle(t))

			if tt.configRaw != "" {
				config := rio.MustYAMLToMap(strings.NewReader(tt.configRaw))
				linter = linter.WithUserConfig(config)
			}

			if tt.ignoreFilesFlag != nil {
				linter = linter.WithIgnore(tt.ignoreFilesFlag)
			}

			result, err := linter.Lint(context.Background(), test.InputPolicy(tt.filename, policy))
			if err != nil {
				t.Fatal(err)
			}

			if len(result.Violations) != len(tt.expViolations) {
				t.Fatalf("expected %d violation, got %d", len(tt.expViolations), len(result.Violations))
			}

			for idx, violation := range result.Violations {
				if violation.Title != tt.expViolations[idx] {
					t.Errorf("expected first violation to be '%s', got %s", tt.expViolations[idx], result.Violations[0].Title)
				}
			}
		})
	}
}

func TestLintWithGoRule(t *testing.T) {
	t.Parallel()

	policy := `package p

 x := true
`

	linter := NewLinter().
		WithAddedBundle(test.GetRegalBundle(t)).WithEnableAll(true)

	result, err := linter.Lint(context.Background(), test.InputPolicy("p.rego", policy))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "opa-fmt" {
		t.Errorf("expected first violation to be 'opa-fmt', got %s", result.Violations[0].Title)
	}
}

func TestLintWithUserConfigGoRuleIgnore(t *testing.T) {
	t.Parallel()

	policy := `package p

 x := true
`

	configRaw := `rules:
  style:
    opa-fmt:
      level: ignore
`

	config := rio.MustYAMLToMap(strings.NewReader(configRaw))

	linter := NewLinter().
		WithUserConfig(config).
		WithAddedBundle(test.GetRegalBundle(t))

	result, err := linter.Lint(context.Background(), test.InputPolicy("p.rego", policy))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Violations) != 0 {
		t.Fatalf("expected no violation, got %d", len(result.Violations))
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
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "acme-corp-package" {
		t.Errorf("expected first violation to be 'acme-corp-package', got %s", result.Violations[0].Title)
	}
}

func TestLintWithCustomRuleAndCustomConfig(t *testing.T) {
	t.Parallel()

	policy := "package p\n"

	configRaw := `rules:
  naming:
    acme-corp-package:
      level: ignore`

	config := rio.MustYAMLToMap(strings.NewReader(configRaw))

	linter := NewLinter().
		WithUserConfig(config).
		WithAddedBundle(test.GetRegalBundle(t)).
		WithCustomRules([]string{filepath.Join("testdata", "custom.rego")})

	result, err := linter.Lint(context.Background(), test.InputPolicy("p.rego", policy))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Violations) != 0 {
		t.Fatalf("expected no violation, got %d", len(result.Violations))
	}
}
