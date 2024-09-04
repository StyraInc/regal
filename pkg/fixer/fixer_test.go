package fixer

import (
	"bytes"
	"context"
	"slices"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"

	"github.com/styrainc/regal/pkg/fixer/fileprovider"
	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/linter"
)

func TestFixer(t *testing.T) {
	t.Parallel()

	policies := map[string][]byte{
		"test/main.rego": []byte(`package test

allow {
true #no space
}

deny = true
`),
	}

	memfp := fileprovider.NewInMemoryFileProvider(policies)

	input, err := memfp.ToInput()
	if err != nil {
		t.Fatalf("failed to create input: %v", err)
	}

	l := linter.NewLinter().
		WithEnableAll(true).
		WithInputModules(&input)

	f := NewFixer()
	f.RegisterFixes(fixes.NewDefaultFixes()...)

	fixReport, err := f.Fix(context.Background(), &l, memfp)
	if err != nil {
		t.Fatalf("failed to fix: %v", err)
	}

	expectedFileFixedViolations := map[string][]string{
		// use-assignment-operator is not expected since use-rego-v1 also addresses this in this example
		"test/main.rego": {"no-whitespace-comment", "use-rego-v1"},
	}
	expectedFileContents := map[string][]byte{
		"test/main.rego": []byte(`package test

import rego.v1

allow := true

# no space

deny := true
`),
	}

	if got, exp := fixReport.TotalFixes(), uint(2); got != exp {
		t.Fatalf("expected %d fixed files, got %d", exp, got)
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

		if !bytes.Equal(content, expectedContent) {
			t.Fatalf("unexpected content for %s:\ngot:\n%s---\nexpected:\n%s---",
				file,
				string(content),
				string(expectedContent))
		}

		// check that the fixed violations are correct
		fxs := fixReport.FixesForFile(file)

		expectedFixes, ok := expectedFileFixedViolations[file]
		if !ok {
			t.Fatalf("unexpected file waas fixed %s", file)
		}

		if len(fxs) != len(expectedFixes) {
			t.Fatalf("unexpected number of fixes for %s:\ngot: %v\nexpected: %v", file, fxs, expectedFixes)
		}

		for _, fx := range fxs {
			if !slices.Contains(expectedFixes, fx.Title) {
				t.Fatalf("expected fixes to contain %s:\ngot: %v", fx.Title, expectedFixes)
			}
		}
	}
}

func TestFixerWithRegisterMandatoryFixes(t *testing.T) {
	t.Parallel()

	policies := map[string][]byte{
		"main.rego": []byte(`package test

allow {
true #no space
}

deny = true
`),
	}

	memfp := fileprovider.NewInMemoryFileProvider(policies)

	input, err := memfp.ToInput()
	if err != nil {
		t.Fatalf("failed to create input: %v", err)
	}

	l := linter.NewLinter().
		WithEnableAll(true).
		WithInputModules(&input)

	f := NewFixer()
	// No fixes are registered here, we are only testing the functionality of
	// RegisterMandatoryFixes
	f.RegisterMandatoryFixes(
		&fixes.Fmt{
			NameOverride: "use-rego-v1",
			OPAFmtOpts: format.Opts{
				RegoVersion: ast.RegoV0CompatV1,
			},
		},
	)

	fixReport, err := f.Fix(context.Background(), &l, memfp)
	if err != nil {
		t.Fatalf("failed to fix: %v", err)
	}

	expectedFileFixedViolations := map[string][]string{
		"main.rego": {"use-rego-v1"},
	}
	expectedFileContents := map[string][]byte{
		// note that since only the rego-v1-format fix is run, the
		// no-whitespace-comment fix is not applied
		"main.rego": []byte(`package test

import rego.v1

allow := true

#no space

deny := true
`),
	}

	if got, exp := fixReport.TotalFixes(), uint(1); got != exp {
		t.Fatalf("expected %d fixed files, got %d", exp, got)
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

		if !bytes.Equal(content, expectedContent) {
			t.Fatalf("unexpected content for %s:\ngot:\n%s---\nexpected:\n%s---",
				file,
				string(content),
				string(expectedContent))
		}

		// check that the fixed violations are correct
		fxs := fixReport.FixesForFile(file)

		expectedFixes, ok := expectedFileFixedViolations[file]
		if !ok {
			t.Fatalf("unexpected file waas fixed %s", file)
		}

		if len(fxs) != len(expectedFixes) {
			t.Fatalf("unexpected number of fixes for %s:\ngot: %v\nexpected: %v", file, fxs, expectedFixes)
		}

		for _, fx := range fxs {
			if !slices.Contains(expectedFixes, fx.Title) {
				t.Fatalf("expected fixes to contain %s:\ngot: %v", fx.Title, expectedFixes)
			}
		}
	}
}
