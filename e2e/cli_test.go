//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/tester"

	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/report"
)

func readProvidedConfig(t *testing.T) config.Config {
	t.Helper()

	cwd := must(os.Getwd)

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
				exp := map[string]any{"files_scanned": zero, "files_failed": zero, "files_skipped": zero, "num_violations": zero}
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
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	td := t.TempDir()

	err := regal(&stdout, &stderr)("lint", td+"/what/ever")
	if exp, act := 1, ExitStatus(err); exp != act {
		t.Errorf("expected exit status %d, got %d", exp, act)
	}

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	if exp, act := "error(s) encountered while linting: errors encountered when reading files to lint: "+
		"failed to filter paths: 1 error occurred during loading: stat "+td+"/what/ever: no such file or directory\n",
		stdout.String(); exp != act {
		t.Errorf("expected stdout %q, got %q", exp, act)
	}
}

func TestLintAllViolations(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cwd := must(os.Getwd)
	cfg := readProvidedConfig(t)

	err := regal(&stdout, &stderr)("lint", "--format", "json", cwd+"/testdata/violations")
	if exp, act := 3, ExitStatus(err); exp != act {
		t.Errorf("expected exit status %d, got %d", exp, act)
	}

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var rep report.Report

	err = json.Unmarshal(stdout.Bytes(), &rep)
	if err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}

	ruleNames := make(map[string]struct{})

	for _, category := range cfg.Rules {
		for ruleName, rule := range category {
			if rule.Level != "ignore" {
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

func TestLintRuleIgnoreFiles(t *testing.T) {
	t.Parallel()

	cwd := must(os.Getwd)
	cfg := readProvidedConfig(t)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	err := regal(&stdout, &stderr)("lint", "--format", "json", "--config-file", cwd+"/testdata/configs/ignore_files_prefer_snake_case.yaml", cwd+"/testdata/violations")

	if exp, act := 3, ExitStatus(err); exp != act {
		t.Errorf("expected exit status %d, got %d", exp, act)
	}

	if exp, act := "", stderr.String(); exp != act {
		t.Errorf("expected stderr %q, got %q", exp, act)
	}

	var rep report.Report

	err = json.Unmarshal(stdout.Bytes(), &rep)
	if err != nil {
		t.Fatalf("expected JSON response, got %v", stdout.String())
	}

	ruleNames := make(map[string]struct{})

	for _, category := range cfg.Rules {
		for ruleName, rule := range category {
			if rule.Level != "ignore" {
				ruleNames[ruleName] = struct{}{}
			}
		}
	}

	violationNames := make(map[string]struct{})

	for _, violation := range rep.Violations {
		violationNames[violation.Title] = struct{}{}
	}

	if available, actual := len(ruleNames), len(violationNames); available != actual+1 {
		t.Errorf("expected one rule to be missing from reported violations, but there were %d out of %d violations reported", actual, available)
	}

	ignoredRuleName := "prefer-snake-case"

	if _, ok := violationNames[ignoredRuleName]; ok {
		t.Errorf("expected violation for rule todo-comments to be ignored")
	}

	// Should be ignored, so exclude from below check
	// also, this has specifically been checked to miss from violationNames above
	delete(ruleNames, ignoredRuleName)

	for ruleName := range ruleNames {
		if _, ok := violationNames[ruleName]; !ok {
			t.Errorf("expected violation for rule %q", ruleName)
		}
	}
}

func TestLintRuleNamingConventionFromCustomCategory(t *testing.T) {
	t.Parallel()

	cwd := must(os.Getwd)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	err := regal(&stdout, &stderr)("lint", "--format", "json", "--config-file", cwd+"/testdata/configs/custom_naming_convention.yaml", cwd+"/testdata/custom_naming_convention")

	if exp, act := 3, ExitStatus(err); exp != act {
		t.Errorf("expected exit status %d, got %d", exp, act)
	}

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
		`Naming convention violation: package name "this.fails" does not match pattern "^acmecorp\\.[a-z_\\.]+$"`,
		`Naming convention violation: rule name "naming_convention_fail" does not match pattern "^_[a-z_]+$|^allow$"`,
	}

	for _, violation := range rep.Violations {
		if !util.Contains(expectedViolations, violation.Description) {
			t.Errorf("unexpected violation: %s", violation.Description)
		}
	}
}

func TestTestRegalBundledRules(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cwd := must(os.Getwd)

	err := regal(&stdout, &stderr)("test", "--format", "json", cwd+"/testdata/custom_rules")

	if exp, act := 0, ExitStatus(err); exp != act {
		t.Errorf("expected exit status %d, got %d", exp, act)
	}

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

	cwd := must(os.Getwd)

	err := regal(&stdout, &stderr)("test", cwd+"/testdata/ast_type_failure")

	if exp, act := 1, ExitStatus(err); exp != act {
		t.Errorf("expected exit status %d, got %d", exp, act)
	}

	expStart := "1 error occurred: "
	expEnd := "rego_type_error: undefined ref: input.foo\n\tinput.foo\n\t      ^\n\t      " +
		"have: \"foo\"\n\t      want (one of): [\"annotations\" \"comments\" \"imports\" \"package\" \"regal\" \"rules\"]\n"

	if !strings.HasPrefix(stdout.String(), expStart) {
		t.Errorf("expected stdout error message starting with %q, got %q", expStart, stdout.String())
	}

	if !strings.HasSuffix(stdout.String(), expEnd) {
		t.Errorf("expected stdout error message ending with %q, got %q", expEnd, stdout.String())
	}
}

func TestCreateNewCustomRuleFromTemplate(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	tmpDir := t.TempDir()

	err := regal(&stdout, &stderr)("new", "rule", "--category", "naming", "--name", "foo-bar-baz", "--output", tmpDir)

	if exp, act := 0, ExitStatus(err); exp != act {
		t.Errorf("expected exit status %d, got %d", exp, act)
	}

	err = regal(&stdout, &stderr)("test", tmpDir)

	if exp, act := 0, ExitStatus(err); exp != act {
		t.Errorf("expected exit status %d, got %d", exp, act)
	}

	if strings.HasPrefix(stdout.String(), "PASS 1/1") {
		t.Errorf("expected stdout to contain PASS 1/1, got %q", stdout.String())
	}
}

func TestCreateNewBuiltinRuleFromTemplate(t *testing.T) {
	t.Parallel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	tmpDir := t.TempDir()

	err := regal(&stdout, &stderr)("new", "rule", "--category", "naming", "--name", "foo-bar-baz", "--output", tmpDir)

	if exp, act := 0, ExitStatus(err); exp != act {
		t.Errorf("expected exit status %d, got %d", exp, act)
	}

	err = regal(&stdout, &stderr)("test", tmpDir)

	if exp, act := 0, ExitStatus(err); exp != act {
		t.Errorf("expected exit status %d, got %d", exp, act)
	}

	if strings.HasPrefix(stdout.String(), "PASS 1/1") {
		t.Errorf("expected stdout to contain PASS 1/1, got %q", stdout.String())
	}
}

func binary() string {
	if b := os.Getenv("REGAL_BIN"); b != "" {
		return b
	}

	return "../regal"
}

func regal(outs ...io.Writer) func(...string) error {
	return func(args ...string) error {
		c := exec.Command(binary(), args...)

		if len(outs) > 0 {
			c.Stdout = outs[0]
		}

		if len(outs) > 1 {
			c.Stderr = outs[0]
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

func must[R any](f func() (R, error)) R {
	var r R

	r, err := f()
	if err != nil {
		log.Fatal(err)
	}

	return r
}
