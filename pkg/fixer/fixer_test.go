package fixer

import (
	"bytes"
	"io"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"

	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/report"
)

func TestFixer_OPAFmtOnly(t *testing.T) {
	t.Parallel()

	rep := &report.Report{
		Violations: []report.Violation{
			{
				Title:       "opa-fmt",
				Description: "File should be formatted with `opa fmt`",
				Category:    "style",
				Location: report.Location{
					File:   "main.rego",
					Row:    1,
					Column: 1,
				},
			},
		},
	}

	fileContents := map[string]io.Reader{
		"main.rego": bytes.NewReader([]byte("package test\nimport rego.v1\nallow := true")),
	}
	expectedFileFixedViolations := map[string]map[string]struct{}{
		"main.rego": {
			"opa-fmt": {},
		},
	}
	expectedFileContents := map[string][]byte{
		"main.rego": []byte(`package test

import rego.v1

allow := true
`),
	}
	f := Fixer{}
	f.RegisterFixes(&fixes.Fmt{})

	fixReport, err := f.Fix(rep, fileContents)
	if err != nil {
		t.Fatalf("failed to fix: %v", err)
	}

	if got, exp := len(fixReport.fileContents), len(expectedFileContents); got != exp {
		t.Fatalf("expected %d fixed files, got %d", exp, got)
	}

	for file, content := range fixReport.fileContents {
		expectedContent, ok := expectedFileContents[file]
		if !ok {
			t.Fatalf("unexpected file %s", file)
		}

		if !bytes.Equal(content, expectedContent) {
			t.Fatalf("unexpected content for %s:\ngot:\n%s---\nexpected:\n%s---",
				file,
				string(content),
				string(expectedContent))
		}
	}

	if got, exp := len(fixReport.fileFixedViolations), len(expectedFileFixedViolations); got != exp {
		t.Fatalf("expected %d fixed files, got %d", exp, got)
	}

	for file, violations := range fixReport.fileFixedViolations {
		expectedViolations, ok := expectedFileFixedViolations[file]
		if !ok {
			t.Fatalf("unexpected file %s", file)
		}

		if got, exp := len(violations), len(expectedViolations); got != exp {
			t.Fatalf("expected %d fixed violations, got %d", exp, got)
		}

		for violation := range violations {
			if _, ok := expectedViolations[violation]; !ok {
				t.Fatalf("unexpected fixed violation %s", violation)
			}
		}
	}

	if got, exp := fixReport.TotalFixes(), 1; got != exp {
		t.Fatalf("expected %d total loadedFixes, got %d", exp, got)
	}
}

func TestFixer_OPAFmtAndRegoV1(t *testing.T) {
	t.Parallel()

	rep := &report.Report{
		Violations: []report.Violation{
			{
				Title: "opa-fmt",
				Location: report.Location{
					File:   "main.rego",
					Row:    1,
					Column: 1,
				},
			},
			{
				Title: "use-rego-v1",
				Location: report.Location{
					File:   "main.rego",
					Row:    1,
					Column: 1,
				},
			},
		},
	}

	fileContents := map[string]io.Reader{
		"main.rego": bytes.NewReader([]byte("package test\nimport rego.v1\nallow := true")),
	}

	expectedFileFixedViolations := map[string]map[string]struct{}{
		"main.rego": {
			"opa-fmt":     {},
			"use-rego-v1": {},
		},
	}
	expectedFileContents := map[string][]byte{
		"main.rego": []byte(`package test

import rego.v1

allow := true
`),
	}

	f := Fixer{}
	f.RegisterFixes(
		&fixes.Fmt{},
		&fixes.Fmt{
			KeyOverride: "use-rego-v1",
			OPAFmtOpts: format.Opts{
				RegoVersion: ast.RegoV0CompatV1,
			},
		},
	)

	fixReport, err := f.Fix(rep, fileContents)
	if err != nil {
		t.Fatalf("failed to fix: %v", err)
	}

	if got, exp := len(fixReport.fileContents), len(expectedFileContents); got != exp {
		t.Fatalf("expected %d fixed files, got %d", exp, got)
	}

	for file, content := range fixReport.fileContents {
		expectedContent, ok := expectedFileContents[file]
		if !ok {
			t.Fatalf("unexpected file %s", file)
		}

		if !bytes.Equal(content, expectedContent) {
			t.Fatalf("unexpected content for %s:\ngot:\n%s---\nexpected:\n%s---",
				file,
				string(content),
				string(expectedContent))
		}
	}

	if got, exp := len(fixReport.fileFixedViolations), len(expectedFileFixedViolations); got != exp {
		t.Fatalf("expected %d fixed files, got %d", exp, got)
	}

	for file, violations := range fixReport.fileFixedViolations {
		expectedViolations, ok := expectedFileFixedViolations[file]
		if !ok {
			t.Fatalf("unexpected file %s", file)
		}

		if got, exp := len(violations), len(expectedViolations); got != exp {
			t.Fatalf("expected %d fixed violations, got %d", exp, got)
		}

		for violation := range violations {
			if _, ok := expectedViolations[violation]; !ok {
				t.Fatalf("unexpected fixed violation %s", violation)
			}
		}
	}

	if got, exp := fixReport.TotalFixes(), 2; got != exp {
		t.Fatalf("expected %d total loadedFixes, got %d", exp, got)
	}
}
