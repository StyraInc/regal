package linter

import (
	"bytes"
	"context"
	"embed"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown"

	"github.com/styrainc/regal/bundle"
	"github.com/styrainc/regal/internal/cache"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/test"
	"github.com/styrainc/regal/internal/testutil"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/report"
	"github.com/styrainc/regal/pkg/rules"
)

func TestLintWithDefaultBundle(t *testing.T) {
	t.Parallel()

	input := test.InputPolicy("p/p.rego", `package p

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

	if *result.Violations[1].Location.Text != "camelCase if {" {
		t.Errorf("expected second violation to be on 'camelCase if {', got %s",
			*result.Violations[1].Location.Text)
	}
}

func TestLintWithUserConfig(t *testing.T) {
	t.Parallel()

	input := test.InputPolicy("p/p.rego", `package p

import rego.v1

r := input.foo[_]
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
		expLevels       []string
		ignoreFilesFlag []string
		rootDir         string
	}{
		{
			name:          "baseline",
			filename:      "p/p.rego",
			expViolations: []string{"top-level-iteration", "rule-shadows-builtin", "opa-fmt"},
		},
		{
			name: "ignore rule",
			userConfig: &config.Config{Rules: map[string]config.Category{
				"bugs":  {"rule-shadows-builtin": config.Rule{Level: "ignore"}},
				"style": {"opa-fmt": config.Rule{Level: "ignore"}},
			}},
			filename:      "p/p.rego",
			expViolations: []string{"top-level-iteration"},
		},
		{
			name: "ignore all",
			userConfig: &config.Config{
				Defaults: config.Defaults{
					Global: config.Default{
						Level: "ignore",
					},
				},
			},
			filename:      "p.rego",
			expViolations: []string{},
		},
		{
			name: "ignore all but bugs",
			userConfig: &config.Config{
				Defaults: config.Defaults{
					Global: config.Default{
						Level: "ignore",
					},
					Categories: map[string]config.Default{
						"bugs": {Level: "error"},
					},
				},
				Rules: map[string]config.Category{
					"bugs": {"rule-shadows-builtin": config.Rule{Level: "ignore"}},
				},
			},
			filename:      "p/p.rego",
			expViolations: []string{"top-level-iteration"},
		},
		{
			name: "ignore style, no global default",
			userConfig: &config.Config{
				Defaults: config.Defaults{
					Categories: map[string]config.Default{
						"bugs":  {Level: "error"},
						"style": {Level: "ignore"},
					},
				},
				Rules: map[string]config.Category{
					"bugs": {"rule-shadows-builtin": config.Rule{Level: "ignore"}},
				},
			},
			filename:      "p/p.rego",
			expViolations: []string{"top-level-iteration"},
		},
		{
			name: "set level to warning",
			userConfig: &config.Config{
				Defaults: config.Defaults{
					Global: config.Default{
						Level: "warning", // will apply to all but style
					},
					Categories: map[string]config.Default{
						"style": {Level: "error"},
					},
				},
				Rules: map[string]config.Category{},
			},
			filename:      "p/p.rego",
			expViolations: []string{"top-level-iteration", "rule-shadows-builtin", "opa-fmt"},
			expLevels:     []string{"warning", "warning", "error"},
		},
		{
			name: "rule level ignore files",
			userConfig: &config.Config{Rules: map[string]config.Category{
				"bugs": {"rule-shadows-builtin": config.Rule{
					Level: "error",
					Ignore: &config.Ignore{
						Files: []string{"p/p.rego"},
					},
				}},
				"style": {"opa-fmt": config.Rule{
					Level: "error",
					Ignore: &config.Ignore{
						Files: []string{"p/p.rego"},
					},
				}},
			}},
			filename:      "p/p.rego",
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
			name: "user config global ignore files with rootDir",
			userConfig: &config.Config{
				Ignore: config.Ignore{
					Files: []string{"foo/*"},
				},
			},
			filename:      "file:///wow/foo/p.rego",
			expViolations: []string{},
			rootDir:       "file:///wow",
		},
		{
			name: "user config global ignore files with rootDir, not ignored",
			userConfig: &config.Config{
				Ignore: config.Ignore{
					Files: []string{"bar/*"},
				},
			},
			filename: "file:///wow/foo/p.rego",
			expViolations: []string{
				"top-level-iteration", "rule-shadows-builtin", "directory-package-mismatch", "opa-fmt",
			},
			rootDir: "file:///wow",
		},
		{
			name:            "CLI flag ignore files",
			filename:        "p.rego",
			expViolations:   []string{},
			ignoreFilesFlag: []string{"p.rego"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := test.InputPolicy(tc.filename, policy)
			linter := NewLinter().
				WithPathPrefix(tc.rootDir).
				WithIgnore(tc.ignoreFilesFlag).
				WithInputModules(&input)

			if tc.userConfig != nil {
				linter = linter.WithUserConfig(*tc.userConfig)
			}

			result := testutil.Must(linter.Lint(context.Background()))(t)

			if len(result.Violations) != len(tc.expViolations) {
				t.Fatalf("expected %d violation, got %d: %v",
					len(tc.expViolations),
					len(result.Violations),
					result.Violations,
				)
			}

			for idx, violation := range result.Violations {
				if violation.Title != tc.expViolations[idx] {
					t.Errorf("expected first violation to be '%s', got %s", tc.expViolations[idx], result.Violations[0].Title)
				}
			}

			if len(tc.expLevels) > 0 {
				for idx, violation := range result.Violations {
					if violation.Level != tc.expLevels[idx] {
						t.Errorf("expected first violation to be '%s', got %s", tc.expLevels[idx], result.Violations[0].Level)
					}
				}
			}
		})
	}
}

func TestLintWithCustomRule(t *testing.T) {
	t.Parallel()

	input := test.InputPolicy("p/p.rego", "package p\n\nimport rego.v1\n")

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

func TestLintWithErrorInEnable(t *testing.T) {
	t.Parallel()

	input := test.InputPolicy("p/p.rego", "package p")

	linter := NewLinter().
		WithCustomRules([]string{filepath.Join("testdata", "custom.rego")}).
		WithEnabledRules("foo").
		WithInputModules(&input)

	_, err := linter.Lint(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if exp, got := "unknown rules: [foo]", err.Error(); !strings.Contains(got, exp) {
		t.Fatalf("expected error to contain %q, got %q", exp, got)
	}
}

//go:embed testdata/*
var testLintWithCustomEmbeddedRulesFS embed.FS

func TestLintWithCustomEmbeddedRules(t *testing.T) {
	t.Parallel()

	input := test.InputPolicy("p/p.rego", "package p\n\nimport rego.v1\n")

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

	input := test.InputPolicy("p/p.rego", "package p\n\nimport rego.v1\n")

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

	mergedConfig := testutil.Must(linter.GetConfig())(t)

	// Since no level was provided, "error" should be inherited from the provided configuration for the rule
	if mergedConfig.Rules["style"]["file-length"].Level != "error" {
		t.Errorf("expected level to be 'error', got %q", mergedConfig.Rules["style"]["file-length"].Level)
	}

	fileLength := mergedConfig.Rules["style"]["file-length"].Extra["max-file-length"]

	// Ensure the extra attributes are still there.
	if fileLength.(int) != 1 { //nolint: forcetypeassert
		t.Errorf("expected max-file-length to be 1, got %d %T", fileLength, fileLength)
	}
}

func TestLintMergedConfigUsesProvidedDefaults(t *testing.T) {
	t.Parallel()

	userConfig := config.Config{
		Defaults: config.Defaults{
			Global: config.Default{
				Level: "ignore",
			},
			Categories: map[string]config.Default{
				"style": {Level: "error"},
				"bugs":  {Level: "warning"},
			},
		},
		Rules: map[string]config.Category{
			"style": {"opa-fmt": config.Rule{Level: "warning"}},
		},
	}

	input := test.InputPolicy("p.rego", `package p`)

	linter := NewLinter().
		WithUserConfig(userConfig).
		WithInputModules(&input)

	mergedConfig := testutil.Must(linter.GetConfig())(t)

	// specifically configured rule should not be affected by the default
	if mergedConfig.Rules["style"]["opa-fmt"].Level != "warning" {
		t.Errorf("expected level to be 'warning', got %q", mergedConfig.Rules["style"]["opa-fmt"].Level)
	}

	// other rule in style should have the default level for the category
	if mergedConfig.Rules["style"]["chained-rule-body"].Level != "error" {
		t.Errorf("expected level to be 'error', got %q", mergedConfig.Rules["style"]["chained-rule-body"].Level)
	}

	// rule in bugs should have the default level for the category
	if mergedConfig.Rules["bugs"]["constant-condition"].Level != "warning" {
		t.Errorf("expected level to be 'warning', got %q", mergedConfig.Rules["bugs"]["constant-condition"].Level)
	}

	// rule in unconfigured category should have the global default level
	if mergedConfig.Rules["imports"]["avoid-importing-input"].Level != "ignore" {
		t.Errorf("expected level to be 'ignore', got %q", mergedConfig.Rules["imports"]["avoid-importing-input"].Level)
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

	if _, err := linter.Lint(context.Background()); err != nil {
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
		t.Fatalf("expected one violation, got %d", len(result.Violations))
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

func TestEnabledRules(t *testing.T) {
	t.Parallel()

	linter := NewLinter().
		WithDisableAll(true).
		WithEnabledRules("opa-fmt", "no-whitespace-comment")

	enabledRules, err := linter.DetermineEnabledRules(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(enabledRules) != 2 {
		t.Fatalf("expected 2 enabled rules, got %d", len(enabledRules))
	}

	if enabledRules[0] != "no-whitespace-comment" {
		t.Errorf("expected first enabled rule to be 'no-whitespace-comment', got %q", enabledRules[0])
	}

	if enabledRules[1] != "opa-fmt" {
		t.Errorf("expected first enabled rule to be 'opa-fmt', got %q", enabledRules[1])
	}
}

func TestEnabledRulesWithConfig(t *testing.T) {
	t.Parallel()

	configFileBs := []byte(`
rules:
  style:
    opa-fmt:
      level: ignore # go rule
  imports:
    unresolved-import: # agg rule
      level: ignore
  idiomatic:
    directory-package-mismatch: # non agg rule
      level: ignore
`)

	var userConfig config.Config

	if err := yaml.Unmarshal(configFileBs, &userConfig); err != nil {
		t.Fatalf("failed to load config: %s", err)
	}

	linter := NewLinter().WithUserConfig(userConfig)

	enabledRules, err := linter.DetermineEnabledRules(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	enabledAggRules, err := linter.DetermineEnabledAggregateRules(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(enabledRules) == 0 {
		t.Fatalf("expected enabledRules, got none")
	}

	if slices.Contains(enabledRules, "directory-package-mismatch") {
		t.Errorf("did not expect directory-package-mismatch to be in enabled rules")
	}

	if slices.Contains(enabledRules, "opa-fmt") {
		t.Errorf("did not expect opa-fmt to be in enabled rules")
	}

	if slices.Contains(enabledAggRules, "unresolved-import") {
		t.Errorf("did not expect unresolved-import to be in enabled aggregate rules")
	}
}

func TestEnabledAggregateRules(t *testing.T) {
	t.Parallel()

	linter := NewLinter().
		WithDisableAll(true).
		WithEnabledRules("opa-fmt", "unresolved-import", "use-assignment-operator")

	enabledRules, err := linter.DetermineEnabledAggregateRules(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(enabledRules) != 1 {
		t.Fatalf("expected 1 enabled rules, got %d", len(enabledRules))
	}

	if enabledRules[0] != "unresolved-import" {
		t.Errorf("expected first enabled rule to be 'unresolved-import', got %q", enabledRules[0])
	}
}

func TestLintWithCollectQuery(t *testing.T) {
	t.Parallel()

	input := test.InputPolicy("p.rego", `package p

import data.foo.bar.unresolved
`)

	linter := NewLinter().
		WithDisableAll(true).
		WithEnabledRules("unresolved-import").
		WithCollectQuery(true).     // needed since we have a single file input
		WithExportAggregates(true). // needed to be able to test the aggregates are set
		WithInputModules(&input)

	result := testutil.Must(linter.Lint(context.Background()))(t)

	if len(result.Aggregates) != 1 {
		t.Fatalf("expected one aggregate, got %d", len(result.Aggregates))
	}

	if _, ok := result.Aggregates["imports/unresolved-import"]; !ok {
		t.Errorf("expected aggregates to contain 'imports/unresolved-import'")
	}
}

func TestLintWithCollectQueryAndAggregates(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"foo.rego": `package foo

import data.unresolved`,
		"bar.rego": `package foo

import data.unresolved`,
		"baz.rego": `package foo

import data.unresolved`,
	}

	allAggregates := make(map[string][]report.Aggregate)

	for file, content := range files {
		input := test.InputPolicy(file, content)

		linter := NewLinter().
			WithDisableAll(true).
			WithEnabledRules("unresolved-import").
			WithCollectQuery(true). // runs collect for a single file input
			WithExportAggregates(true).
			WithInputModules(&input)

		result := testutil.Must(linter.Lint(context.Background()))(t)

		for k, aggs := range result.Aggregates {
			allAggregates[k] = append(allAggregates[k], aggs...)
		}
	}

	linter := NewLinter().
		WithDisableAll(true).
		WithEnabledRules("unresolved-import").
		WithAggregates(allAggregates)

	result := testutil.Must(linter.Lint(context.Background()))(t)

	if len(result.Violations) != 3 {
		t.Fatalf("expected one violation, got %d", len(result.Violations))
	}

	foundFiles := []string{}

	for _, v := range result.Violations {
		if v.Title != "unresolved-import" {
			t.Errorf("unexpected title: %s", v.Title)
		}

		foundFiles = append(foundFiles, v.Location.File)
	}

	slices.Sort(foundFiles)

	if exp, got := []string{"bar.rego", "baz.rego", "foo.rego"}, foundFiles; !slices.Equal(exp, got) {
		t.Fatalf("unexpected files: %v", got)
	}
}

// 1034816959 ns/op	2910685336 B/op	55738401 allocs/op
// ...
func BenchmarkRegalLintingItself(b *testing.B) {
	linter := NewLinter().
		WithInputPaths([]string{"../../bundle"}).
		WithBaseCache(cache.NewBaseCache()).
		WithEnableAll(true)

	b.ResetTimer()
	b.ReportAllocs()

	var err error

	var rep report.Report

	for range b.N {
		rep, err = linter.Lint(context.Background())
		if err != nil {
			b.Fatal(err)
		}
	}

	if len(rep.Violations) != 0 {
		_ = rep.Violations
	}
}

// 268203979 ns/op	501748262 B/op	 9724107 allocs/op
// 258819136 ns/op	497016108 B/op	 9632445 allocs/op
// 170215493 ns/op	497115590 B/op	 9228106 allocs/op
// 155879726 ns/op	444934614 B/op	 8362362 allocs/op
// ...
func BenchmarkRegalNoEnabledRules(b *testing.B) {
	linter := NewLinter().
		WithInputPaths([]string{"../../bundle"}).
		WithBaseCache(cache.NewBaseCache()).
		WithDisableAll(true)

	b.ResetTimer()
	b.ReportAllocs()

	var err error

	var rep report.Report

	for range b.N {
		rep, err = linter.Lint(context.Background())
		if err != nil {
			b.Fatal(err)
		}
	}

	if len(rep.Violations) != 0 {
		_ = rep.Violations
	}
}

// Runs a separate benchmark for each rule in the bundle. Note that this will take *several* minutes to run,
// meaning you do NOT want to do this more than occasionally. You may however find it helpful to use this with
// a single, or handful of rules to get a better idea of how long they take to run, and relative to each other.
// The reason why this is currently so slow is that we parse + compile everything for each rule, which is really
// something we should fix: https://github.com/StyraInc/regal/issues/1394
func BenchmarkEachRule(b *testing.B) {
	conf, err := config.LoadConfigWithDefaultsFromBundle(&bundle.LoadedBundle, nil)
	if err != nil {
		b.Fatal(err)
	}

	for _, category := range conf.Rules {
		for ruleName := range category {
			// Uncomment / modify this to benchmark specific rule(s) only
			//
			// if ruleName != "metasyntactic-variable" {
			// 	continue
			// }
			b.Run(ruleName, func(b *testing.B) {
				linter := NewLinter().
					WithInputPaths([]string{"../../bundle"}).
					WithBaseCache(cache.NewBaseCache()).
					WithDisableAll(true).
					WithEnabledRules(ruleName)

				b.ResetTimer()
				b.ReportAllocs()

				var err error

				var rep report.Report

				for range b.N {
					rep, err = linter.Lint(context.Background())
					if err != nil {
						b.Fatal(err)
					}
				}

				if len(rep.Violations) != 0 {
					_ = rep.Violations
				}
			})
		}
	}
}
