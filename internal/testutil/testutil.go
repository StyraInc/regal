package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/styrainc/regal/pkg/report"
	rutil "github.com/styrainc/regal/pkg/roast/util"
)

func Must[T any](x T, err error) func(t *testing.T) T {
	return func(t *testing.T) T {
		t.Helper()

		if err != nil {
			t.Fatal(err)
		}

		return x
	}
}

func TempDirectoryOf(t *testing.T, files map[string]string) string {
	t.Helper()

	tmpDir := t.TempDir()

	for file, contents := range files {
		path := filepath.Join(tmpDir, file)

		MustMkdirAll(t, filepath.Dir(path))
		MustWriteFile(t, path, []byte(contents))
	}

	return tmpDir
}

func MustMkdirAll(t *testing.T, path ...string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(path...), 0o755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

func MustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}

	return contents
}

func MustWriteFile(t *testing.T, path string, contents []byte) {
	t.Helper()

	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func MustRemove(t *testing.T, path string) {
	t.Helper()

	if err := os.Remove(path); err != nil {
		t.Fatalf("failed to remove file %s: %v", path, err)
	}
}

func AssertNumViolations(t *testing.T, num int, rep report.Report) {
	t.Helper()

	if rep.Summary.NumViolations != num {
		t.Errorf("expected %d violations, got %d", num, rep.Summary.NumViolations)
	}
}

func ViolationTitles(rep report.Report) *rutil.Set[string] {
	titles := make([]string, len(rep.Violations))
	for i := range rep.Violations {
		titles[i] = rep.Violations[i].Title
	}

	return rutil.NewSet(titles...)
}

func AssertOnlyViolations(t *testing.T, rep report.Report, expected ...string) {
	t.Helper()

	violationNames := ViolationTitles(rep)

	if violationNames.Size() != len(expected) {
		t.Errorf("expected %d violations, got %d: %v", len(expected), violationNames.Size(), violationNames.Items())
	}

	for _, name := range expected {
		if !violationNames.Contains(name) {
			t.Errorf("expected violation for rule %q, but it was not found", name)
		}
	}
}

func AssertContainsViolations(t *testing.T, rep report.Report, expected ...string) {
	t.Helper()

	violationNames := ViolationTitles(rep)

	for _, name := range expected {
		if !violationNames.Contains(name) {
			t.Errorf("expected violation for rule %q, but it was not found", name)
		}
	}
}

func AssertNotContainsViolations(t *testing.T, rep report.Report, unexpected ...string) {
	t.Helper()

	violationNames := ViolationTitles(rep)
	if violationNames.Contains(unexpected...) {
		t.Errorf("expected no violations for rules %v, but found: %v", unexpected, violationNames.Items())
	}
}

func RemoveIgnoreErr(paths ...string) func() {
	return func() {
		for _, path := range paths {
			_ = os.Remove(path)
		}
	}
}
