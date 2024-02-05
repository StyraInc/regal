//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/tester"

	"github.com/styrainc/regal/internal/testutil"
	"github.com/styrainc/regal/internal/util"
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

	err = yaml.Unmarshal(bs, &cfg)
	if err != nil {
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
		tc := tc
		t.Run(tc.format, func(t *testing.T) {
			t.Parallel()

			out := bytes.Buffer{}

			err := regal(&out)("lint", "--format", tc.format, t.TempDir())
			if err != nil {
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

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	td := t.TempDir()

	err := regal(&stdout, &stderr)("lint", td+filepath.FromSlash("/what/ever"))

	expectExitCode(t, err, 1, &stdout, &stderr)

	if exp, act := "", stdout.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	if exp, act := "error(s) encountered while linting: errors encountered when reading files to lint: "+
		"failed to filter paths:\nstat "+td+filepath.FromSlash("/what/ever")+": no such file or directory\n",
		stderr.String(); exp != act {
		t.Errorf("expected stdout %q, got %q", exp, act)
	}
}

func TestLintAllViolations(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)
	cfg := readProvidedConfig(t)

	err := regal(&stdout, &stderr)("lint", "--format", "json", cwd+filepath.FromSlash("/testdata/violations"))

	expectExitCode(t, err, 3, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var rep report.Report

	err = json.Unmarshal(stdout.Bytes(), &rep)
	if err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}

	ruleNames := make(map[string]struct{})

	excludedRules := map[string]struct{}{
		"implicit-future-keywords": {},
		"use-if":                   {},
		"use-contains":             {},
	}

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

func TestLintNotRegoV1Violations(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)("lint", "--format", "json", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/not_rego_v1.yaml"),
		cwd+filepath.FromSlash("/testdata/not_rego_v1"))

	expectExitCode(t, err, 3, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var rep report.Report

	err = json.Unmarshal(stdout.Bytes(), &rep)
	if err != nil {
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

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

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

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	err := regal(&stdout, &stderr)("lint", "--format", "json", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/ignore_files_prefer_snake_case.yaml"),
		cwd+filepath.FromSlash("/testdata/violations"))

	expectExitCode(t, err, 3, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var rep report.Report

	err = json.Unmarshal(stdout.Bytes(), &rep)
	if err != nil {
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

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

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

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	err := regal(&stdout, &stderr)("lint", "--format", "json", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/custom_naming_convention.yaml"),
		cwd+filepath.FromSlash("/testdata/custom_naming_convention"))

	expectExitCode(t, err, 3, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var rep report.Report

	err = json.Unmarshal(stdout.Bytes(), &rep)
	if err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}

	if rep.Summary.NumViolations != 2 {
		t.Errorf("expected 2 violations, got %d", rep.Summary.NumViolations)
	}

	expectedViolations := []string{
		`Naming convention violation: package name "this.fails" does not match pattern '^acmecorp\.[a-z_\.]+$'`,
		`Naming convention violation: rule name "naming_convention_fail" does not match pattern '^_[a-z_]+$|^allow$'`,
	}

	for _, violation := range rep.Violations {
		if !util.Contains(expectedViolations, violation.Description) {
			t.Errorf("unexpected violation: %s", violation.Description)
		}
	}
}

func TestAggregatesAreCollectedAndUsed(t *testing.T) {
	t.Parallel()
	cwd := testutil.Must(os.Getwd())(t)
	basedir := cwd + filepath.FromSlash("/testdata/aggregates")

	t.Run("two policies — no violations expected", func(t *testing.T) {
		stdout := bytes.Buffer{}
		stderr := bytes.Buffer{}

		err := regal(&stdout, &stderr)("lint", "--format", "json", "--rules",
			basedir+filepath.FromSlash("/rules/custom_rules_using_aggregates.rego"),
			basedir+filepath.FromSlash("/two_policies"))

		expectExitCode(t, err, 0, &stdout, &stderr)

		if exp, act := "", stderr.String(); exp != act {
			t.Errorf("expected stderr %q, got %q", exp, act)
		}
	})

	t.Run("single policy — no aggregate violations expected", func(t *testing.T) {
		stdout := bytes.Buffer{}
		stderr := bytes.Buffer{}

		err := regal(&stdout, &stderr)("lint", "--format", "json", "--rules",
			basedir+filepath.FromSlash("/rules/custom_rules_using_aggregates.rego"),
			basedir+filepath.FromSlash("/two_policies/policy_1.rego"))

		expectExitCode(t, err, 0, &stdout, &stderr)

		if exp, act := "", stderr.String(); exp != act {
			t.Errorf("expected stderr %q, got %q", exp, act)
		}
	})

	t.Run("three policies - violation expected", func(t *testing.T) {
		stdout := bytes.Buffer{}
		stderr := bytes.Buffer{}
		// By sending a single file to the command, we skip the aggregates computation, so we expect one violation
		err := regal(&stdout, &stderr)("lint", "--format", "json", "--rules",
			basedir+filepath.FromSlash("/rules/custom_rules_using_aggregates.rego"),
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
}

func TestTestRegalBundledBundle(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)("test", "--format", "json", cwd+filepath.FromSlash("/../bundle"))

	expectExitCode(t, err, 0, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var res []tester.Result

	err = json.Unmarshal(stdout.Bytes(), &res)
	if err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}
}

func TestTestRegalBundledRules(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)("test", "--format", "json",
		cwd+filepath.FromSlash("/testdata/custom_rules"))

	expectExitCode(t, err, 0, &stdout, &stderr)

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var res []tester.Result

	err = json.Unmarshal(stdout.Bytes(), &res)
	if err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}
}

func TestTestRegalTestWithExtendedASTTypeChecking(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)

	err := regal(&stdout, &stderr)("test", cwd+filepath.FromSlash("/testdata/ast_type_failure"))

	expectExitCode(t, err, 1, &stdout, &stderr)

	expStart := "1 error occurred: "
	expEnd := "rego_type_error: undefined ref: input.foo\n\tinput.foo\n\t      ^\n\t      " +
		"have: \"foo\"\n\t      want (one of): [\"annotations\" \"comments\" \"imports\" \"package\" \"regal\" \"rules\"]\n"

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

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
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

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
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

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)

	// No violations from the built-in configuration in the policy provided, but
	// the user --config-file changes the max-file-length to 1, so this should fail
	err := regal(&stdout, &stderr)("lint", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/rule_without_level.yaml"),
		cwd+filepath.FromSlash("/testdata/custom_naming_convention"))

	expectExitCode(t, err, 3, &stdout, &stderr)
}

func TestLintWithCustomCapabilitiesAndUnmetRequirement(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)

	// Test that the custom-has-key rule is skipped due to the custom capabilities provided where we
	// use OPA v0.46.0 as a target (the `object.keys` built-in function was introduced in v0.47.0)
	err := regal(&stdout, &stderr)("lint", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/opa_v46_capabilities.yaml"),
		cwd+filepath.FromSlash("/testdata/capabilities/custom_has_key.rego"))

	// This is only an informative warning — command should not fail
	expectExitCode(t, err, 0, &stdout, &stderr)

	expectOut := "1 file linted. No violations found. 2 rules skipped:\n" +
		"- custom-has-key-construct: Missing capability for built-in function `object.keys`\n" +
		"- use-rego-v1: Missing capability for `import rego.v1`\n\n"

	if stdout.String() != expectOut {
		t.Errorf("expected %q, got %q", expectOut, stdout.String())
	}
}

func TestLintWithCustomCapabilitiesAndUnmetRequirementMultipleFiles(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)

	// Test that the custom-has-key rule is skipped due to the custom capabilities provided where we
	// use OPA v0.46.0 as a target (the `object.keys` built-in function was introduced in v0.47.0)
	err := regal(&stdout, &stderr)("lint", "--config-file",
		cwd+filepath.FromSlash("/testdata/configs/opa_v46_capabilities.yaml"),
		cwd+filepath.FromSlash("/testdata/capabilities/"))

	// This is only an informative warning — command should not fail
	expectExitCode(t, err, 0, &stdout, &stderr)

	expectOut := "2 files linted. No violations found. 2 rules skipped:\n" +
		"- custom-has-key-construct: Missing capability for built-in function `object.keys`\n" +
		"- use-rego-v1: Missing capability for `import rego.v1`\n\n"

	if stdout.String() != expectOut {
		t.Errorf("expected %q, got %q", expectOut, stdout.String())
	}
}

func TestLintPprof(t *testing.T) {
	t.Parallel()

	const pprofFile = "clock.pprof"

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cwd := testutil.Must(os.Getwd())(t)

	t.Cleanup(func() {
		os.Remove(pprofFile)
	})

	err := regal(&stdout, &stderr)("lint", "--pprof", "clock", cwd+filepath.FromSlash("/testdata/violations"))

	expectExitCode(t, err, 3, &stdout, &stderr)

	_, err = os.Stat(pprofFile)
	if err != nil {
		t.Fatalf("expected to find %s, got error: %v", pprofFile, err)
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
