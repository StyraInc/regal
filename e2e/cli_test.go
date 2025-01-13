//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/v1/tester"

	"github.com/styrainc/regal/internal/testutil"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/report"
)

func readProvidedConfig(t *testing.T) config.Config {
	t.Helper()

	cwd := testutil.Must(os.Getwd())(t)

	configPath := filepath.Join(cwd, "..", "bundle", "regal", "config", "provided", "data.yaml")
	bs, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config from %q: %v", configPath, err)
	}

	var cfg config.Config
	if err = yaml.Unmarshal(bs, &cfg); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	return cfg
}

func TestCLIUsage(t *testing.T) {
	t.Parallel()

	if err := regal()(); err != nil {
		t.Fatal(err)
	}
}

func TestLintEmptyDir(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		format string
		check  func(*testing.T, *bytes.Buffer)
	}{
		{
			format: "pretty",
			check: func(t *testing.T, out *bytes.Buffer) {
				t.Helper()
				if exp, act := "0 files linted. No violations found.\n", out.String(); exp != act {
					t.Errorf("output: expected %q, got %q", exp, act)
				}
			},
		},
		{
			format: "compact",
			check: func(t *testing.T, out *bytes.Buffer) {
				t.Helper()
				if exp, act := "\n", out.String(); exp != act {
					t.Errorf("output: expected %q, got %q", exp, act)
				}
			},
		},
		{
			format: "json",
			check: func(t *testing.T, out *bytes.Buffer) {
				t.Helper()
				s := struct {
					Violations []string       `json:"violations"`
					Summary    map[string]any `json:"summary"`
				}{}
				if err := json.NewDecoder(out).Decode(&s); err != nil {
					t.Fatal(err)
				}
				if exp, act := 0, len(s.Violations); exp != act {
					t.Errorf("violations: expected %d, got %d", exp, act)
				}
				zero := float64(0)
				exp := map[string]any{"files_scanned": zero, "files_failed": zero, "rules_skipped": zero, "num_violations": zero}
				if diff := cmp.Diff(exp, s.Summary); diff != "" {
					t.Errorf("unexpected summary (-want, +got):\n%s", diff)
				}
			},
		},
	} {
		t.Run(tc.format, func(t *testing.T) {
			t.Parallel()

			out := bytes.Buffer{}
			if err := regal(&out)("lint", "--format", tc.format, t.TempDir()); err != nil {
				t.Fatalf("%v %[1]T", err)
			}

			tc.check(t, &out)
		})
	}
}

func TestLintNonExistentDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows as the error message is different")
	}

	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	td := t.TempDir()

	err := regal(&stdout, &stderr)("lint", td+filepath.FromSlash("/what/ever"))

	expectExitCode(t, err, 1, &stdout, &stderr)

	if exp, act := "", stdout.String(); exp != act {
		t.Errorf("expected stdout %q, got %q", exp, act)
	}

	if exp, act := "error(s) encountered while linting: errors encountered when reading files to lint: "+
		"failed to filter paths:\nstat "+td+filepath.FromSlash("/what/ever")+": no such file or directory\n",
		stderr.String(); exp != act {
		t.Errorf("expected stderr\n%q,\ngot\n%q", exp, act)
	}
}

func TestLintProposeToRunFix(t *testing.T) {
	t.Parallel()
	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)

	// using a test rego file that only yields a few violations
	err := regal(&stdout, &stderr)(
		"lint",
		"--config-file", filepath.Join(cwd, "e2e_conf.yaml"),
		cwd+filepath.FromSlash("/testdata/v0/rule_named_if.rego"))

	expectExitCode(t, err, 3, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Fatalf("expected stderr %q, got %q", exp, act)
	}

	act := strings.Split(stdout.String(), "\n")
	act = act[len(act)-5:]
	exp := []string{
		"1 file linted. 2 violations found.",
		"",
		"Hint: 2/2 violations can be automatically fixed (directory-package-mismatch, opa-fmt)",
		"      Run regal fix --help for more details.",
		"",
	}
	if diff := cmp.Diff(act, exp); diff != "" {
		t.Errorf("unexpected stdout trailer: (-want, +got):\n%s", diff)
	}
}

func TestLintV1Violations(t *testing.T) {
	t.Parallel()
	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)(
		"lint",
		"--format", "json",
		"--config-file", filepath.Join(cwd, "e2e_conf.yaml"),
		cwd+filepath.FromSlash("/testdata/violations"),
	)

	expectExitCode(t, err, 3, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var rep report.Report
	if err = json.Unmarshal(stdout.Bytes(), &rep); err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}

	ruleNames := make(map[string]struct{})

	excludedRules := map[string]struct{}{
		"implicit-future-keywords": {},
		"use-if":                   {},
		"use-contains":             {},
		"internal-entrypoint":      {},
		"file-length":              {},
		"rule-named-if":            {},
		"use-rego-v1":              {},
		"deprecated-builtin":       {},
		"import-shadows-import":    {},
	}

	cfg := readProvidedConfig(t)

	for _, category := range cfg.Rules {
		for ruleName, rule := range category {
			if _, isExcluded := excludedRules[ruleName]; !isExcluded && rule.Level != "ignore" {
				ruleNames[ruleName] = struct{}{}
			}
		}
	}

	// Note that some violations occur more than one time.
	violationNames := make(map[string]struct{})

	for _, violation := range rep.Violations {
		violationNames[violation.Title] = struct{}{}
	}

	if len(ruleNames) != len(violationNames) {
		for ruleName := range ruleNames {
			if _, ok := violationNames[ruleName]; !ok {
				t.Errorf("expected violation for rule %q", ruleName)
			}
		}
	}
}

func TestLintV0NoRegoV1ImportViolations(t *testing.T) {
	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)("lint", "--format", "json", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/v0.yaml"),
		cwd+filepath.FromSlash("/testdata/v0/"))

	expectExitCode(t, err, 3, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var rep report.Report
	if err = json.Unmarshal(stdout.Bytes(), &rep); err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}

	expected := map[string]struct{}{
		"implicit-future-keywords": {},
		"use-if":                   {},
		"use-contains":             {},
	}

	// Note that some violations occur more than one time.
	violationNames := make(map[string]struct{})

	for _, violation := range rep.Violations {
		violationNames[violation.Title] = struct{}{}
	}

	if len(expected) != len(violationNames) {
		for ruleName := range expected {
			if _, ok := violationNames[ruleName]; !ok {
				t.Errorf("expected violation for rule %q", ruleName)
			}
		}
	}
}

func TestLintV0WithRegoV1ImportViolations(t *testing.T) {
	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)(
		"lint", "--format", "json",
		"--config-file", cwd+filepath.FromSlash("/testdata/configs/v0-with-import-rego-v1.yaml"),
		cwd+filepath.FromSlash("/testdata/v0/"))

	expectExitCode(t, err, 3, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var rep report.Report
	if err = json.Unmarshal(stdout.Bytes(), &rep); err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}

	expected := map[string]struct{}{
		"use-if":        {},
		"use-contains":  {},
		"use-rego-v1":   {},
		"rule-named-if": {},
	}

	// Note that some violations occur more than one time.
	violationNames := make(map[string]struct{})

	for _, violation := range rep.Violations {
		violationNames[violation.Title] = struct{}{}
	}

	if len(expected) != len(violationNames) {
		for ruleName := range expected {
			if _, ok := violationNames[ruleName]; !ok {
				t.Errorf("expected violation for rule %q", ruleName)
			}
		}
	}
}

func TestLintFailsNonExistentConfigFile(t *testing.T) {
	t.Parallel()

	var expected string
	switch runtime.GOOS {
	case "windows":
		expected = "The system cannot find the file specified"
	default:
		expected = "no such file or directory"
	}

	cwd := testutil.Must(os.Getwd())(t)

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}

	err := regal(&stdout, &stderr)("lint", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/non_existent_test_file.yaml"),
		cwd+filepath.FromSlash("/testdata/violations"))

	expectExitCode(t, err, 1, &stdout, &stderr)

	if !strings.Contains(stderr.String(), expected) {
		t.Errorf("expected stderr to print, got %q", stderr.String())
	}
}

func TestLintRuleIgnoreFiles(t *testing.T) {
	t.Parallel()

	cwd := testutil.Must(os.Getwd())(t)

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}

	err := regal(&stdout, &stderr)("lint", "--format", "json", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/ignore_files_prefer_snake_case.yaml"),
		cwd+filepath.FromSlash("/testdata/violations"))

	expectExitCode(t, err, 3, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var rep report.Report
	if err = json.Unmarshal(stdout.Bytes(), &rep); err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}

	violationNames := make(map[string]struct{})

	for _, violation := range rep.Violations {
		violationNames[violation.Title] = struct{}{}
	}

	if _, ok := violationNames["prefer-snake-case"]; ok {
		t.Errorf("did not expect violation for rule %q as it is ignored", "prefer-snake-case")
	}
}

func TestLintWithDebugOption(t *testing.T) {
	t.Parallel()

	cwd := testutil.Must(os.Getwd())(t)
	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}

	err := regal(&stdout, &stderr)("lint", "--debug", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/ignore_files_prefer_snake_case.yaml"),
		cwd+filepath.FromSlash("/testdata/violations"))

	expectExitCode(t, err, 3, &stdout, &stderr)

	if !strings.Contains(stderr.String(), "rules:") {
		t.Errorf("expected stderr to print configuration, got %q", stderr.String())
	}
}

func TestLintRuleNamingConventionFromCustomCategory(t *testing.T) {
	t.Parallel()

	cwd := testutil.Must(os.Getwd())(t)
	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}

	err := regal(&stdout, &stderr)("lint", "--format", "json", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/custom_naming_convention.yaml"),
		cwd+filepath.FromSlash("/testdata/custom_naming_convention"))

	expectExitCode(t, err, 3, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var rep report.Report
	if err = json.Unmarshal(stdout.Bytes(), &rep); err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}

	if rep.Summary.NumViolations != 2 {
		t.Errorf("expected 2 violations, got %d", rep.Summary.NumViolations)
	}

	expectedViolations := []string{
		`Naming convention violation: package name "custom_naming_convention" does not match pattern '^acmecorp\.[a-z_\.]+$'`,
		`Naming convention violation: rule name "naming_convention_fail" does not match pattern '^_[a-z_]+$|^allow$'`,
	}

	for _, violation := range rep.Violations {
		if !slices.Contains(expectedViolations, violation.Description) {
			t.Errorf("unexpected violation: %s", violation.Description)
		}
	}
}

func TestAggregatesAreCollectedAndUsed(t *testing.T) {
	t.Parallel()
	cwd := testutil.Must(os.Getwd())(t)
	basedir := cwd + filepath.FromSlash("/testdata/aggregates")

	t.Run("two policies — no violations expected", func(t *testing.T) {
		stdout, stderr := bytes.Buffer{}, bytes.Buffer{}

		err := regal(&stdout, &stderr)("lint", "--format", "json",
			basedir+filepath.FromSlash("/custom/regal/rules/testcase/aggregates/custom_rules_using_aggregates.rego"),
			basedir+filepath.FromSlash("/two_policies"))

		expectExitCode(t, err, 0, &stdout, &stderr)

		if exp, act := "", stderr.String(); exp != act {
			t.Errorf("expected stderr %q, got %q", exp, act)
		}
	})

	t.Run("single policy — no aggregate violations expected", func(t *testing.T) {
		stdout, stderr := bytes.Buffer{}, bytes.Buffer{}

		err := regal(&stdout, &stderr)("lint", "--format", "json", "--rules",
			basedir+filepath.FromSlash("/custom/regal/rules/testcase/aggregates/custom_rules_using_aggregates.rego"),
			basedir+filepath.FromSlash("/two_policies/policy_1.rego"))

		expectExitCode(t, err, 0, &stdout, &stderr)

		if exp, act := "", stderr.String(); exp != act {
			t.Errorf("expected stderr %q, got %q", exp, act)
		}
	})

	t.Run("three policies - violation expected", func(t *testing.T) {
		stdout, stderr := bytes.Buffer{}, bytes.Buffer{}

		err := regal(&stdout, &stderr)("lint", "--format", "json",
			"--config-file", filepath.Join(cwd, "e2e_conf.yaml"),
			"--rules",
			basedir+filepath.FromSlash("/custom/regal/rules/testcase/aggregates/custom_rules_using_aggregates.rego"),
			basedir+filepath.FromSlash("/three_policies"))

		expectExitCode(t, err, 3, &stdout, &stderr)

		if exp, act := "", stderr.String(); exp != act {
			t.Errorf("expected stderr %q, got %q", exp, act)
		}

		var rep report.Report

		if err = json.Unmarshal(stdout.Bytes(), &rep); err != nil {
			t.Fatalf("expected JSON response, got %v", stdout.String())
		}

		if rep.Summary.NumViolations != 1 {
			t.Errorf("expected 1 violation, got %d", rep.Summary.NumViolations)
		}
	})

	t.Run("custom policy where nothing aggregate is a violation", func(t *testing.T) {
		stdout, stderr := bytes.Buffer{}, bytes.Buffer{}

		err := regal(&stdout, &stderr)("lint", "--format", "json",
			"--config-file", filepath.Join(cwd, "e2e_conf.yaml"),
			"--rules",
			basedir+filepath.FromSlash("/custom/regal/rules/testcase/empty_aggregate/"),
			basedir+filepath.FromSlash("/two_policies"))

		expectExitCode(t, err, 3, &stdout, &stderr)

		if exp, act := "", stderr.String(); exp != act {
			t.Errorf("expected stderr %q, got %q", exp, act)
		}

		var rep report.Report

		if err = json.Unmarshal(stdout.Bytes(), &rep); err != nil {
			t.Fatalf("expected JSON response, got %v", stdout.String())
		}

		if rep.Summary.NumViolations != 1 {
			t.Errorf("expected 1 violation, got %d", rep.Summary.NumViolations)
		}
	})
}

func TestLintAggregateIgnoreDirective(t *testing.T) {
	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)(
		"lint",
		"--config-file", filepath.Join(cwd, "e2e_conf.yaml"),
		"--format",
		"json",
		cwd+filepath.FromSlash("/testdata/aggregates/ignore_directive"),
	)

	expectExitCode(t, err, 3, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var rep report.Report

	if err = json.Unmarshal(stdout.Bytes(), &rep); err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}

	if rep.Summary.NumViolations != 2 {
		t.Errorf("expected 2 violations, got %d", rep.Summary.NumViolations)
	}

	if rep.Summary.NumViolations == 0 {
		t.Fatal("expected violations, got none")
	}

	if rep.Violations[0].Title != "no-defined-entrypoint" {
		t.Errorf("expected violation 'no-defined-entrypoint', got %q", rep.Violations[0].Title)
	}

	if rep.Violations[1].Title != "unresolved-import" {
		t.Errorf("expected violation 'unresolved-import', got %q", rep.Violations[1].Title)
	}

	// ensure that it's the file without the ignore directive that has the violation
	if !strings.HasSuffix(rep.Violations[1].Location.File, "second.rego") {
		t.Errorf("expected violation in second.rego, got %q", rep.Violations[1].Location.File)
	}
}

func TestTestRegalBundledBundle(t *testing.T) {
	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)("test", "--format", "json", cwd+filepath.FromSlash("/../bundle"))

	expectExitCode(t, err, 0, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var res []tester.Result

	if err = json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}
}

func TestTestRegalBundledRules(t *testing.T) {
	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)("test", "--format", "json", cwd+filepath.FromSlash("/testdata/custom_rules"))

	expectExitCode(t, err, 0, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var res []tester.Result
	if err = json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}
}

func TestTestRegalTestWithExtendedASTTypeChecking(t *testing.T) {
	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)("test", cwd+filepath.FromSlash("/testdata/ast_type_failure"))

	expectExitCode(t, err, 1, &stdout, &stderr)

	expStart := "1 error occurred: "
	expEnd := "rego_type_error: undefined ref: input.foo\n\tinput.foo\n\t      ^\n\t      " +
		"have: \"foo\"\n\t      want (one of): [\"comments\" \"imports\" \"package\" \"regal\" \"rules\"]\n"

	if !strings.HasPrefix(stderr.String(), expStart) {
		t.Errorf("expected stdout error message starting with %q, got %q", expStart, stderr.String())
	}

	if !strings.HasSuffix(stderr.String(), expEnd) {
		t.Errorf("expected stdout error message ending with %q, got %q", expEnd, stderr.String())
	}
}

// Both of the template creating tests are skipped on Windows for the time being,
// as the "regal test" command fails with a "no such file or directory" error, even
// though the files are seemingly created. Will need to look into this, but these
// tests are not critical.

func TestCreateNewCustomRuleFromTemplate(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("temporarily skipping this test on Windows")
	}

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	tmpDir := t.TempDir()

	expectExitCode(t, regal(&stdout, &stderr)(
		"new", "rule", "--category", "naming", "--name", "foo-bar-baz", "--output", tmpDir,
	), 0, &stdout, &stderr)

	stdout.Reset()
	stderr.Reset()

	expectExitCode(t, regal(&stdout, &stderr)("test", tmpDir), 0, &stdout, &stderr)

	if strings.HasPrefix(stdout.String(), "PASS 1/1") {
		t.Errorf("expected stdout to contain PASS 1/1, got %q", stdout.String())
	}
}

func TestCreateNewBuiltinRuleFromTemplate(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("temporarily skipping this test on Windows")
	}

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	tmpDir := t.TempDir()

	expectExitCode(t, regal(&stdout, &stderr)(
		"new", "rule", "--category", "naming", "--name", "foo-bar-baz", "--output", tmpDir,
	), 0, &stdout, &stderr)

	stdout.Reset()
	stderr.Reset()

	expectExitCode(t, regal(&stdout, &stderr)("test", tmpDir), 0, &stdout, &stderr)

	if strings.HasPrefix(stdout.String(), "PASS 1/1") {
		t.Errorf("expected stdout to contain PASS 1/1, got %q", stdout.String())
	}
}

func TestMergeRuleConfigWithoutLevel(t *testing.T) {
	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	cwd := testutil.Must(os.Getwd())(t)

	// No violations from the built-in configuration in the policy provided, but
	// the user --config-file changes the max-file-length to 1, so this should fail
	err := regal(&stdout, &stderr)("lint", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/rule_without_level.yaml"),
		cwd+filepath.FromSlash("/testdata/custom_naming_convention"))

	expectExitCode(t, err, 3, &stdout, &stderr)
}

func TestConfigDefaultingWithDisableDirective(t *testing.T) {
	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)(
		"lint",
		"--disable-category=testing",
		"--config-file",
		cwd+filepath.FromSlash("/testdata/configs/defaulting.yaml"),
		cwd+filepath.FromSlash("/testdata/defaulting"),
	)

	// ignored by flag ignore directive
	if strings.Contains(stdout.String(), "print-or-trace-call") {
		t.Errorf("expected stdout to not contain print-or-trace-call")
		t.Log("stdout:\n", stdout.String())
	}

	// ignored by config
	if strings.Contains(stdout.String(), "opa-fmt") {
		t.Errorf("expected stdout to not contain print-or-trace-call")
		t.Log("stdout:\n", stdout.String())
	}

	// this error should not be ignored
	if !strings.Contains(stdout.String(), "top-level-iteration") {
		t.Errorf("expected stdout to contain top-level-iteration")
		t.Log("stdout:\n", stdout.String())
	}

	expectExitCode(t, err, 3, &stdout, &stderr)
}

func TestConfigDefaultingWithEnableDirective(t *testing.T) {
	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)(
		"lint",
		"--enable-all",
		"--config-file",
		cwd+filepath.FromSlash("/testdata/configs/defaulting.yaml"),
		cwd+filepath.FromSlash("/testdata/defaulting"),
	)

	// re-enabled by flag enable directive
	if !strings.Contains(stdout.String(), "print-or-trace-call") {
		t.Errorf("expected stdout to contain print-or-trace-call")
		t.Log("stdout:\n", stdout.String())
	}

	// re-enabled by flag enable directive
	if !strings.Contains(stdout.String(), "opa-fmt") {
		t.Errorf("expected stdout to contain opa-fmt")
		t.Log("stdout:\n", stdout.String())
	}

	// this error should not be ignored
	if !strings.Contains(stdout.String(), "top-level-iteration") {
		t.Errorf("expected stdout to contain top-level-iteration")
		t.Log("stdout:\n", stdout.String())
	}

	expectExitCode(t, err, 3, &stdout, &stderr)
}

func TestLintWithCustomCapabilitiesAndUnmetRequirement(t *testing.T) {
	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	cwd := testutil.Must(os.Getwd())(t)

	// Test that the custom-has-key rule is skipped due to the custom capabilities provided where we
	// use OPA v0.46.0 as a target (the `object.keys` built-in function was introduced in v0.47.0)
	err := regal(&stdout, &stderr)("lint", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/opa_v46_capabilities.yaml"),
		cwd+filepath.FromSlash("/testdata/capabilities/custom_has_key.rego"))

	// This is only an informative warning — command should not fail
	expectExitCode(t, err, 0, &stdout, &stderr)

	expectOut := "1 file linted. No violations found. 3 rules skipped:\n" +
		"- custom-has-key-construct: Missing capability for built-in function `object.keys`\n" +
		"- use-strings-count: Missing capability for built-in function `strings.count`\n" +
		"- use-rego-v1: Missing capability for `import rego.v1`\n\n"

	if stdout.String() != expectOut {
		t.Errorf("expected %q, got %q", expectOut, stdout.String())
	}
}

func TestLintWithCustomCapabilitiesAndUnmetRequirementMultipleFiles(t *testing.T) {
	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	cwd := testutil.Must(os.Getwd())(t)

	// Test that the custom-has-key rule is skipped due to the custom capabilities provided where we
	// use OPA v0.46.0 as a target (the `object.keys` built-in function was introduced in v0.47.0)
	err := regal(&stdout, &stderr)("lint", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/opa_v46_capabilities.yaml"),
		cwd+filepath.FromSlash("/testdata/capabilities/"))

	// This is only an informative warning — command should not fail
	expectExitCode(t, err, 0, &stdout, &stderr)

	expectOut := "2 files linted. No violations found. 3 rules skipped:\n" +
		"- custom-has-key-construct: Missing capability for built-in function `object.keys`\n" +
		"- use-strings-count: Missing capability for built-in function `strings.count`\n" +
		"- use-rego-v1: Missing capability for `import rego.v1`\n\n"

	if stdout.String() != expectOut {
		t.Errorf("expected %q, got %q", expectOut, stdout.String())
	}
}

func TestLintPprof(t *testing.T) {
	t.Parallel()

	const pprofFile = "clock.pprof"

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	cwd := testutil.Must(os.Getwd())(t)

	t.Cleanup(func() {
		_ = os.Remove(pprofFile)
	})

	err := regal(&stdout, &stderr)(
		"lint",
		// this overrides the ignore directives for e2e loaded from the config file
		"--ignore-files=none",
		"--pprof", "clock",
		cwd+filepath.FromSlash("/testdata/violations"),
	)

	expectExitCode(t, err, 3, &stdout, &stderr)

	if _, err = os.Stat(pprofFile); err != nil {
		t.Fatalf("expected to find %s, got error: %v", pprofFile, err)
	}
}

func TestFix(t *testing.T) {
	t.Parallel()

	stdout, stderr := bytes.Buffer{}, bytes.Buffer{}
	td := t.TempDir()

	initialState := map[string]string{
		".regal/config.yaml": `
project:
  rego-version: 1
  roots:
    - path: v0
      rego-version: 0
    - path: v1
      rego-version: 0
`,
		"foo/main.rego": `package wow

#comment

allow if {
	true
}
`,
		"foo/main_test.rego": `package wow_test

test_allow if {
	true
}
`,
		"foo/foo.rego": `package foo

# present and correct

allow if {
	input.admin
}
`,
		"bar/main.rego": `package wow["foo-bar"].baz
`,
		"bar/main_test.rego": `package wow["foo-bar"].baz_test
test_allow if {
	true
}
`,
		"v0/main.rego": `package v0

#comment
allow { true }
`,
		"v1/main.rego": `package v1

#comment
allow if { true }
`,
		"unrelated.txt": `foobar`,
	}

	for file, content := range initialState {
		mustWriteToFile(t, filepath.Join(td, file), content)
	}

	// --force is required to make the changes when there is no git repo
	err := regal(&stdout, &stderr)(
		"fix",
		"--force",
		filepath.Join(td, "foo"),
		filepath.Join(td, "bar"),
		filepath.Join(td, "v0"),
		filepath.Join(td, "v1"),
	)

	// 0 exit status is expected as all violations should have been fixed
	expectExitCode(t, err, 0, &stdout, &stderr)

	exp := fmt.Sprintf(`12 fixes applied:
In project root: %[1]s
bar/main.rego -> wow/foo-bar/baz/main.rego:
- directory-package-mismatch

bar/main_test.rego -> wow/foo-bar/baz/main_test.rego:
- directory-package-mismatch
- opa-fmt

foo/main.rego -> wow/main.rego:
- directory-package-mismatch
- opa-fmt
- no-whitespace-comment

foo/main_test.rego -> wow/main_test.rego:
- directory-package-mismatch
- opa-fmt


In project root: %[1]s/v0
main.rego:
- opa-fmt
- no-whitespace-comment

In project root: %[1]s/v1
main.rego:
- opa-fmt
- no-whitespace-comment
`, td)

	if act := stdout.String(); exp != act {
		t.Errorf("expected stdout:\n%s\ngot:\n%s", exp, act)
	}

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	expectedState := map[string]string{
		"foo/foo.rego": `package foo

# present and correct

allow if {
	input.admin
}
`,
		"wow/foo-bar/baz/main.rego": `package wow["foo-bar"].baz
`,
		"wow/foo-bar/baz/main_test.rego": `package wow["foo-bar"].baz_test

test_allow := true
`,
		"wow/main.rego": `package wow

# comment

allow := true
`,
		"wow/main_test.rego": `package wow_test

test_allow := true
`,
		"v0/main.rego": `package v0

import rego.v1

# comment
allow := true
`,
		"v1/main.rego": `package v1

# comment
allow := true
`,
		"unrelated.txt": `foobar`,
	}

	for file, expectedContent := range expectedState {
		bs := testutil.Must(os.ReadFile(filepath.Join(td, file)))(t)

		if act := string(bs); expectedContent != act {
			t.Errorf("expected %s contents:\n%s\ngot\n%s", file, expectedContent, act)
		}
	}

	// foo is not removed as it contains a correct file
	expectedMissingDirs := []string{"bar"}
	for _, dir := range expectedMissingDirs {
		if _, err := os.Stat(filepath.Join(td, dir)); err == nil {
			t.Errorf("expected directory %q to have been removed", dir)
		}
	}
}

func TestFixWithConflicts(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	td := t.TempDir()

	initialState := map[string]string{
		".regal/config.yaml": "", // needed to find the root in the right place
		// this file is in the correct location
		"foo/foo.rego": `package foo

import rego.v1
`,
		// this file should be at foo/foo.rego, but that file already exists
		"quz/foo.rego": `package foo

import rego.v1
`,
		// these three files should all be at bar/bar.rego, but they cannot all be moved there
		"foo/bar.rego": `package bar

import rego.v1
`,
		"baz/bar.rego": `package bar

import rego.v1
`,
		"bax/foo/wow/bar.rego": `package bar

import rego.v1
`,
	}

	for file, content := range initialState {
		mustWriteToFile(t, filepath.Join(td, file), content)
	}

	// --force is required to make the changes when there is no git repo
	err := regal(&stdout, &stderr)("fix", "--force", td)

	// 0 exit status is expected as all violations should have been fixed
	expectExitCode(t, err, 1, &stdout, &stderr)

	expStdout := fmt.Sprintf(`Source file conflicts:
In project root: %[1]s
Cannot overwrite existing file: foo/foo.rego
- quz/foo.rego

Many to one conflicts:
In project root: %[1]s
Cannot move multiple files to: bar/bar.rego
- bax/foo/wow/bar.rego
- baz/bar.rego
- foo/bar.rego
`, td)

	if act := stdout.String(); expStdout != act {
		t.Errorf("expected stdout:\n%s\ngot\n%s", expStdout, act)
	}

	for file, expectedContent := range initialState {
		bs := testutil.Must(os.ReadFile(filepath.Join(td, file)))(t)

		if act := string(bs); expectedContent != act {
			t.Errorf("expected %s contents:\n%s\ngot\n%s", file, expectedContent, act)
		}
	}
}

func TestFixWithConflictRenaming(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	td := t.TempDir()

	initialState := map[string]string{
		".regal/config.yaml": "", // needed to find the root in the right place
		// this file is in the correct location
		"foo/foo.rego": `package foo

import rego.v1
`,
		// this file is in the correct location
		"foo/foo_test.rego": `package foo_test

import rego.v1
`,
		// this file should be at foo/foo.rego, but that file already exists
		"quz/foo.rego": `package foo

import rego.v1
`,
		// this file should be at foo/foo_test.rego, but that file already exists
		"quz/foo_test.rego": `package foo_test

import rego.v1
`,
		// this file should be at bar/bar.rego and is not a conflict
		"foo/bar.rego": `package bar

import rego.v1
`,
	}

	for file, content := range initialState {
		mustWriteToFile(t, filepath.Join(td, file), content)
	}

	// --force is required to make the changes when there is no git repo
	// --conflict=rename will rename inbound files when there is a conflict
	err := regal(&stdout, &stderr)("fix", "--force", "--on-conflict=rename", td)

	// 0 exit status is expected as all violations should have been fixed
	expectExitCode(t, err, 0, &stdout, &stderr)

	expStdout := fmt.Sprintf(`3 fixes applied:
In project root: %[1]s
foo/bar.rego -> bar/bar.rego:
- directory-package-mismatch
quz/foo.rego -> foo/foo_1.rego:
- directory-package-mismatch
quz/foo_test.rego -> foo/foo_1_test.rego:
- directory-package-mismatch
`, td)

	if act := stdout.String(); expStdout != act {
		t.Errorf("expected stdout:\n%s\ngot\n%s", expStdout, act)
	}

	expectedState := map[string]string{
		".regal/config.yaml": "", // needed to find the root in the right place
		// unchanged
		"foo/foo.rego": `package foo

import rego.v1
`,
		// renamed to permit its new location
		"foo/foo_1.rego": `package foo

import rego.v1
`,
		// renamed to permit its new location
		"foo/foo_1_test.rego": `package foo_test

import rego.v1
`,
		// unchanged
		"bar/bar.rego": `package bar

import rego.v1
`,
	}

	for file, expectedContent := range expectedState {
		bs, err := os.ReadFile(filepath.Join(td, file))
		if err != nil {
			t.Errorf("failed to read %s: %v", file, err)

			continue
		}

		if act := string(bs); expectedContent != act {
			t.Errorf("expected %s contents:\n%s\ngot\n%s", file, expectedContent, act)
		}
	}

	expectedMissing := []string{
		"quz/foo.rego",
		"quz/foo_test.rego",
	}

	for _, file := range expectedMissing {
		if _, err := os.Stat(filepath.Join(td, file)); err == nil {
			t.Errorf("expected %s to have been removed", file)
		}
	}
}

// verify fix for https://github.com/StyraInc/regal/issues/1082
func TestLintAnnotationCustomAttributeMultipleItems(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)(
		"lint",
		"--config-file", filepath.Join(cwd, "e2e_conf.yaml"),
		"--disable=directory-package-mismatch",
		filepath.Join(cwd, "testdata", "bugs", "issue_1082.rego"),
	)

	expectExitCode(t, err, 0, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	if exp, act := "1 file linted. No violations found.\n", stdout.String(); exp != act {
		t.Errorf("expected stdout %q, got %q", exp, act)
	}
}

func binary() string {
	var location string
	if runtime.GOOS == "windows" {
		location = "../regal.exe"
	} else {
		location = "../regal"
	}

	if b := os.Getenv("REGAL_BIN"); b != "" {
		location = b
	}

	if _, err := os.Stat(location); errors.Is(err, os.ErrNotExist) {
		log.Fatal("regal binary not found — make sure to run go build before running the e2e tests")
	} else if err != nil {
		log.Fatal(err)
	}

	return location
}

func regal(outs ...io.Writer) func(...string) error {
	return func(args ...string) error {
		c := exec.Command(binary(), args...)

		if len(outs) > 0 {
			c.Stdout = outs[0]
		}

		if len(outs) > 1 {
			c.Stderr = outs[1]
		}

		return c.Run() //nolint:wrapcheck // We're in tests. This is fine.
	}
}

type exitStatus interface {
	ExitStatus() int
}

// ExitStatus returns the exit status of the error if it is an exec.ExitError
// or if it implements ExitStatus() int.
// 0 if it is nil or panics if it is a different error.
func ExitStatus(err error) int {
	switch e := err.(type) { //nolint:errorlint // We know the errors that can happen here, the switch is enough.
	case nil:
		return 0
	case exitStatus:
		return e.ExitStatus()
	case *exec.ExitError:
		if ex, ok := e.Sys().(exitStatus); ok {
			return ex.ExitStatus()
		}
	}

	panic("unreachable")
}

func expectExitCode(t *testing.T, err error, exp int, stdout *bytes.Buffer, stderr *bytes.Buffer) {
	t.Helper()

	if act := ExitStatus(err); exp != act {
		t.Errorf("expected exit status %d, got %d\nstdout: %s\nstderr: %s",
			exp, act, stdout.String(), stderr.String())
	}
}

func mustWriteToFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create directory %s: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write to %s: %v", path, err)
	}
}
