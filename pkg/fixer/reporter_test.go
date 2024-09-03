package fixer

import (
	"bytes"
	"testing"

	"github.com/styrainc/regal/pkg/fixer/fixes"
)

func TestPrettyReporterOutput(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer

	reporter := NewPrettyReporter(&buffer)

	report := NewReport()

	report.AddFileFix("/workspace/bundle1/policy1.rego", fixes.FixResult{
		Title: "rego-v1",
		Root:  "/workspace/bundle1",
	})
	report.AddFileFix("/workspace/bundle2/policy1.rego", fixes.FixResult{
		Title: "rego-v1",
		Root:  "/workspace/bundle2",
	})
	report.AddFileFix("/workspace/bundle1/policy1.rego", fixes.FixResult{
		Title: "directory-package-mismatch",
		Root:  "/workspace/bundle1",
	})
	report.AddFileFix("/workspace/bundle2/policy1.rego", fixes.FixResult{
		Title: "directory-package-mismatch",
		Root:  "/workspace/bundle2",
	})
	report.AddFileFix("/workspace/bundle1/policy3.rego", fixes.FixResult{
		Title: "no-whitespace-comment",
		Root:  "/workspace/bundle1",
	})
	report.AddFileFix("/workspace/bundle2/policy3.rego", fixes.FixResult{
		Title: "use-assignment-operator",
		Root:  "/workspace/bundle2",
	})

	report.MergeFixes(
		"/workspace/bundle1/main/policy1.rego",
		"/workspace/bundle1/policy1.rego",
	)

	err := report.RegisterOldPathForFile(
		"/workspace/bundle1/main/policy1.rego",
		"/workspace/bundle1/policy1.rego",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	report.MergeFixes(
		"/workspace/bundle2/lib/policy2.rego",
		"/workspace/bundle2/policy1.rego",
	)

	err = report.RegisterOldPathForFile(
		"/workspace/bundle2/lib/policy2.rego",
		"/workspace/bundle2/policy1.rego",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = reporter.Report(report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `6 fixes applied:
In project root: /workspace/bundle1
policy1.rego -> main/policy1.rego:
- rego-v1
- directory-package-mismatch
policy3.rego:
- no-whitespace-comment

In project root: /workspace/bundle2
policy1.rego -> lib/policy2.rego:
- rego-v1
- directory-package-mismatch
policy3.rego:
- use-assignment-operator
`

	if got := buffer.String(); got != expected {
		t.Fatalf("unexpected output:\nexpected:\n%s\ngot:\n%s", expected, got)
	}
}
