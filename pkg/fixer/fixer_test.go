package fixer

import (
	"context"
	"slices"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/fixer/fileprovider"
	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/linter"
	"github.com/styrainc/regal/pkg/report"
)

func TestFixer(t *testing.T) {
	t.Parallel()

	policies := map[string]string{
		"/root/main/main.rego": `package test

allow if {
true #no space
}
deny = true
`,
	}

	memfp := fileprovider.NewInMemoryFileProvider(policies)

	input, err := memfp.ToInput(map[string]ast.RegoVersion{
		"/root/main": ast.RegoV1,
	})
	if err != nil {
		t.Fatalf("failed to create input: %v", err)
	}

	l := linter.NewLinter().
		WithEnableAll(true).
		WithInputModules(&input)

	f := NewFixer()
	f.RegisterFixes(fixes.NewDefaultFixes()...)
	f.RegisterRoots("/root")
	f.SetRegoVersionsMap(map[string]ast.RegoVersion{
		"/root/main": ast.RegoV1,
	})

	fixReport, err := f.Fix(context.Background(), &l, memfp)
	if err != nil {
		t.Fatalf("failed to fix: %v", err)
	}

	expectedFileFixedViolations := map[string][]string{
		// use-assigment-operator is correct in formatting so does not appear.
		"/root/test/main.rego": {"directory-package-mismatch", "no-whitespace-comment", "opa-fmt"},
	}
	expectedFileContents := map[string]string{
		"/root/test/main.rego": `package test

allow := true

# no space

deny := true
`,
	}

	if got, exp := fixReport.TotalFixes(), uint(3); got != exp {
		t.Fatalf("expected a total of %d fixes, got %d", exp, got)
	}

	fpFiles, err := memfp.List()
	if err != nil {
		t.Fatalf("failed to list files: %v", err)
	}

	for _, file := range fpFiles {
		// check that the content is correct
		expectedContent, ok := expectedFileContents[file]
		_, moved := fixReport.movedFiles[file]

		if !ok && !moved {
			t.Fatalf("unexpected file found in resulting file provider %s", file)
		}

		content, err := memfp.Get(file)
		if err != nil {
			t.Fatalf("failed to get file %s: %v", file, err)
		}

		if content != expectedContent {
			t.Fatalf(
				"unexpected content for %s:\ngot:\n%s---\nexpected:\n%s---",
				file,
				content,
				expectedContent,
			)
		}

		// check that the fixed violations are correct
		fxs := fixReport.FixesForFile(file)

		var fixes []string
		for _, fx := range fxs {
			fixes = append(fixes, fx.Title)
		}

		expectedFixes, ok := expectedFileFixedViolations[file]
		if !ok {
			t.Fatalf("unexpected file was fixed %s", file)
		}

		slices.Sort(expectedFixes)
		slices.Sort(fixes)

		if !slices.Equal(expectedFixes, fixes) {
			t.Fatalf("unexpected fixes for %s:\ngot: %v\nexpected: %v", file, fixes, expectedFixes)
		}
	}
}

func TestFixViolations(t *testing.T) {
	t.Parallel()

	// targeted specific fix for a given violation
	violations := []report.Violation{
		{
			Title:    "directory-package-mismatch",
			Location: report.Location{File: "root/main.rego"},
		},
	}

	policies := map[string]string{
		"root/main.rego": `package foo.bar

allow := true
`,
		// file in correct place
		"root/foo/bar/main.rego": `package foo.bar

allow := true
`,
	}

	memfp := fileprovider.NewInMemoryFileProvider(policies)

	f := NewFixer()
	f.SetOnConflictOperation(OnConflictRename)
	f.RegisterFixes(fixes.NewDefaultFixes()...)

	fixReport, err := f.FixViolations(violations, memfp, &config.Config{})
	if err != nil {
		t.Fatalf("failed to fix: %v", err)
	}

	expectedFileFixedViolations := map[string][]string{
		"root/main.rego":           {"directory-package-mismatch"},
		"root/foo/bar/main_1.rego": {"directory-package-mismatch"},
		"root/foo/bar/main.rego":   {}, // no fixes
	}
	expectedFileContents := map[string]string{
		// old file yet to be deleted
		"root/main.rego": `package foo.bar

allow := true
`,
		"root/foo/bar/main_1.rego": `package foo.bar

allow := true
`,
		// file in correct place
		"root/foo/bar/main.rego": `package foo.bar

allow := true
`,
	}

	if got, exp := fixReport.TotalFixes(), uint(2); got != exp {
		t.Fatalf("expected %d fixes, got %d", exp, got)
	}

	fpFiles, err := memfp.List()
	if err != nil {
		t.Fatalf("failed to list files: %v", err)
	}

	for _, file := range fpFiles {
		// check that the content is correct
		expectedContent, ok := expectedFileContents[file]
		if !ok {
			t.Fatalf("unexpected file %s", file)
		}

		content, err := memfp.Get(file)
		if err != nil {
			t.Fatalf("failed to get file %s: %v", file, err)
		}

		if content != expectedContent {
			t.Fatalf("unexpected content for %s:\ngot:\n%s---\nexpected:\n%s---",
				file,
				content,
				expectedContent)
		}

		fxs := fixReport.FixesForFile(file)

		// check that the fixed violations are correct
		var fixes []string
		for _, fx := range fxs {
			fixes = append(fixes, fx.Title)
		}

		expectedFixes, ok := expectedFileFixedViolations[file]
		if !ok {
			t.Fatalf("unexpected file was fixed %s", file)
		}

		slices.Sort(expectedFixes)
		slices.Sort(fixes)

		if !slices.Equal(expectedFixes, fixes) {
			t.Fatalf("unexpected fixes for %s:\ngot: %v\nexpected: %v", file, fixes, expectedFixes)
		}
	}
}
