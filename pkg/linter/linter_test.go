package linter

import (
	"bytes"
	"context"
	"embed"
	"path/filepath"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown"

	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/test"
	"github.com/styrainc/regal/internal/testutil"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/rules"
)

func TestLintWithDefaultBundle(t *testing.T) {
	t.Parallel()

	input := test.InputPolicy("p.rego", `package p

import rego.v1

# TODO: fix this
camelCase if {
	input.one == 1
	input.two == 2
}
`)

	linter := NewLinter().WithEnableAll(true).WithInputModules(&input)

	result := testutil.Must(linter.Lint(context.Background()))(t)

	if len(result.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "todo-comment" {
		t.Errorf("expected first violation to be 'todo-comments', got %s", result.Violations[0].Title)
	}

	if result.Violations[0].Location.Row != 5 {
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

	if result.Violations[1].Location.Row != 6 {
		t.Errorf("expected second violation to be on line 4, got %d", result.Violations[1].Location.Row)
	}

	if result.Violations[1].Location.Column != 1 {
		t.Errorf("expected second violation to be on column 1, got %d", result.Violations[1].Location.Column)
	}

	if *result.Violations[1].Location.Text != "camelCase if {" {
		t.Errorf("expected second violation to be on 'camelCase if {', got %s",
			*result.Violations[1].Location.Text)
	}
}

func TestLintWithUserConfig(t *testing.T) {
	t.Parallel()

	input := test.InputPolicy("p.rego", `package p

import rego.v1

boo := input.hoo[_]

or := 1
`)

	userConfig := config.Config{
		Rules: map[string]config.Category{
			"bugs": {"rule-shadows-builtin": config.Rule{Level: "ignore"}},
		},
	}

	linter := NewLinter().WithUserConfig(userConfig).WithInputModules(&input)

	result := testutil.Must(linter.Lint(context.Background()))(t)

	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d - violations: %v", len(result.Violations), result.Violations)
	}

	if result.Violations[0].Title != "top-level-iteration" {
		t.Errorf("expected first violation to be 'top-level-iteration', got %s", result.Violations[0].Title)
	}
}

func TestLintWithUserConfigTable(t *testing.T) {
	t.Parallel()

	policy := `package p

import rego.v1

boo := input.hoo[_]

 opa_fmt := "fail"

or := 1
`
	tests := []struct {
		name            string
		userConfig      *config.Config
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
			userConfig: &config.Config{Rules: map[string]config.Category{
				"bugs":  {"rule-shadows-builtin": config.Rule{Level: "ignore"}},
				"style": {"opa-fmt": config.Rule{Level: "ignore"}},
			}},
			filename:      "p.rego",
			expViolations: []string{"top-level-iteration"},
		},
		{
			name: "rule level ignore files",
			userConfig: &config.Config{Rules: map[string]config.Category{
				"bugs": {"rule-shadows-builtin": config.Rule{
					Level: "error",
					Ignore: &config.Ignore{
						Files: []string{"p.rego"},
					},
				}},
				"style": {"opa-fmt": config.Rule{
					Level: "error",
					Ignore: &config.Ignore{
						Files: []string{"p.rego"},
					},
				}},
			}},
			filename:      "p.rego",
			expViolations: []string{"top-level-iteration"},
		},
		{
			name: "user config global ignore files",
			userConfig: &config.Config{
				Ignore: config.Ignore{
					Files: []string{"p.rego"},
				},
			},
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

			linter := NewLinter()

			if tt.userConfig != nil {
				linter = linter.WithUserConfig(*tt.userConfig)
			}

			if tt.ignoreFilesFlag != nil {
				linter = linter.WithIgnore(tt.ignoreFilesFlag)
			}

			input := test.InputPolicy(tt.filename, policy)

			linter = linter.WithInputModules(&input)

			result := testutil.Must(linter.Lint(context.Background()))(t)

			if len(result.Violations) != len(tt.expViolations) {
				t.Fatalf("expected %d violation, got %d: %v",
					len(tt.expViolations),
					len(result.Violations),
					result.Violations,
				)
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

	input := test.InputPolicy("p.rego", `package p
		import rego.v1

 		x := true
	`)

	linter := NewLinter().
		WithEnableAll(true).
		WithInputModules(&input)

	result := testutil.Must(linter.Lint(context.Background()))(t)

	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "opa-fmt" {
		t.Errorf("expected first violation to be 'opa-fmt', got %s", result.Violations[0].Title)
	}
}

func TestLintWithUserConfigGoRuleIgnore(t *testing.T) {
	t.Parallel()

	userConfig := config.Config{Rules: map[string]config.Category{
		"style": {"opa-fmt": config.Rule{Level: "ignore"}},
	}}

	input := test.InputPolicy("p.rego", `package p
		import rego.v1

	 	x := true
	`)

	linter := NewLinter().
		WithUserConfig(userConfig).
		WithInputModules(&input)

	result := testutil.Must(linter.Lint(context.Background()))(t)

	if len(result.Violations) != 0 {
		t.Fatalf("expected no violation, got %d", len(result.Violations))
	}
}

func TestLintWithCustomRule(t *testing.T) {
	t.Parallel()

	input := test.InputPolicy("p.rego", "package p\n\nimport rego.v1\n")

	linter := NewLinter().
		WithCustomRules([]string{filepath.Join("testdata", "custom.rego")}).
		WithInputModules(&input)

	result := testutil.Must(linter.Lint(context.Background()))(t)

	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "acme-corp-package" {
		t.Errorf("expected first violation to be 'acme-corp-package', got %s", result.Violations[0].Title)
	}
}

//go:embed testdata/*
var testLintWithCustomEmbeddedRulesFS embed.FS

func TestLintWithCustomEmbeddedRules(t *testing.T) {
	t.Parallel()

	input := test.InputPolicy("p.rego", "package p\n\nimport rego.v1\n")

	linter := NewLinter().
		WithCustomRulesFromFS(testLintWithCustomEmbeddedRulesFS, "testdata").
		WithInputModules(&input)

	result := testutil.Must(linter.Lint(context.Background()))(t)

	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].Title != "acme-corp-package" {
		t.Errorf("expected first violation to be 'acme-corp-package', got %s", result.Violations[0].Title)
	}
}

func TestLintWithCustomRuleAndCustomConfig(t *testing.T) {
	t.Parallel()

	input := test.InputPolicy("p.rego", "package p\n\nimport rego.v1\n")

	userConfig := config.Config{Rules: map[string]config.Category{
		"naming": {"acme-corp-package": config.Rule{Level: "ignore"}},
	}}

	linter := NewLinter().
		WithUserConfig(userConfig).
		WithCustomRules([]string{filepath.Join("testdata", "custom.rego")}).
		WithInputModules(&input)

	result := testutil.Must(linter.Lint(context.Background()))(t)

	if len(result.Violations) != 0 {
		t.Fatalf("expected no violation, got %d", len(result.Violations))
	}
}

func TestLintMergedConfigInheritsLevelFromProvided(t *testing.T) {
	t.Parallel()

	// Note that the user configuration does not provide a level
	userConfig := config.Config{Rules: map[string]config.Category{
		"style": {"file-length": config.Rule{Extra: config.ExtraAttributes{"max-file-length": 1}}},
	}}

	input := test.InputPolicy("p.rego", `package p

	x := 1
	`)

	linter := NewLinter().
		WithUserConfig(userConfig).
		WithInputModules(&input)

	mergedConfig := testutil.Must(linter.mergedConfig())(t)

	// Since no level was provided, "error" should be inherited from the provided configuration for the rule
	if mergedConfig.Rules["style"]["file-length"].Level != "error" {
		t.Errorf("expected level to be 'error', got %q", mergedConfig.Rules["style"]["file-length"].Level)
	}

	fileLength := mergedConfig.Rules["style"]["file-length"].Extra["max-file-length"]

	// Ensure the extra attributes are still there.
	if fileLength.(int) != 1 { // nolint: forcetypeassert
		t.Errorf("expected max-file-length to be 1, got %d %T", fileLength, fileLength)
	}
}

func TestLintWithPrintHook(t *testing.T) {
	t.Parallel()

	input := test.InputPolicy("p.rego", `package p`)

	var bb bytes.Buffer

	linter := NewLinter().
		WithCustomRules([]string{filepath.Join("testdata", "printer.rego")}).
		WithPrintHook(topdown.NewPrintHook(&bb)).
		WithInputModules(&input)

	_, err := linter.Lint(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if bb.String() != "p.rego\n" {
		t.Errorf("expected print hook to print file name 'p.rego' and newline, got %q", bb.String())
	}
}

func TestLintWithAggregateRule(t *testing.T) {
	t.Parallel()

	policies := make(map[string]string)
	policies["foo.rego"] = `package foo
		import data.bar

		default allow := false
	`
	policies["bar.rego"] = `package bar
		import data.foo.allow
	`

	modules := make(map[string]*ast.Module)

	for filename, content := range policies {
		modules[filename] = parse.MustParseModule(content)
	}

	input := rules.NewInput(policies, modules)

	linter := NewLinter().
		WithDisableAll(true).
		WithEnabledRules("prefer-package-imports").
		WithInputModules(&input)

	result := testutil.Must(linter.Lint(context.Background()))(t)

	if len(result.Violations) != 1 {
		t.Fatalf("expected no violation, got %d", len(result.Violations))
	}

	violation := result.Violations[0]

	if violation.Title != "prefer-package-imports" {
		t.Errorf("expected violation to be 'prefer-package-imports', got %q", violation.Title)
	}

	if violation.Location.Row != 2 {
		t.Errorf("expected violation to be on line 2, got %d", violation.Location.Row)
	}

	if violation.Location.Column != 3 {
		t.Errorf("expected violation to be on column 3, got %d", violation.Location.Column)
	}

	if *violation.Location.Text != "\t\timport data.foo.allow" {
		t.Errorf("expected violation to be on 'import data.foo.allow', got %q", *violation.Location.Text)
	}
}
