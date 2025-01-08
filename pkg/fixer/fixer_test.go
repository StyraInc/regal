package fixer

import (
	"bytes"
	"context"
	"os"
	"slices"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/format"
	"github.com/open-policy-agent/opa/v1/topdown"

	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/fixer/fileprovider"
	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/linter"
)

func TestFixer(t *testing.T) {
	t.Parallel()

	policies := map[string][]byte{
		"test/main.rego": []byte(`package test

allow = true #no space
`),
	}

	memfp := fileprovider.NewInMemoryFileProvider(policies)

	input, err := memfp.ToInput()
	if err != nil {
		t.Fatalf("failed to create input: %v", err)
	}

	l := linter.NewLinter().
		WithEnableAll(true).
		WithInputModules(&input).
		WithPrintHook(topdown.NewPrintHook(os.Stderr)).
		WithUserConfig(config.Config{
			Capabilities: &config.Capabilities{
				Features: []string{"rego_v1_import"},
			},
		})

	f := NewFixer()
	f.RegisterFixes(
		[]fixes.Fix{
			&fixes.UseAssignmentOperator{},
			&fixes.NoWhitespaceComment{},
			&fixes.DirectoryPackageMismatch{},
		}...,
	)

	fixReport, err := f.Fix(context.Background(), &l, memfp)
	if err != nil {
		t.Fatalf("failed to fix: %v", err)
	}

	expectedFileFixedViolations := map[string][]string{
		"test/main.rego": {"no-whitespace-comment", "use-assignment-operator"},
	}
	expectedFileContents := map[string][]byte{
		"test/main.rego": []byte(`package test

allow := true # no space
`),
	}

	if got, exp := fixReport.TotalFixes(), uint(2); got != exp {
		t.Fatalf("expected a total of %d fixes, got %d", exp, got)
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
		WithInputModules(&input).
		WithUserConfig(config.Config{
			Capabilities: &config.Capabilities{
				Features: []string{"rego_v1_import"},
			},
		})

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

		if !bytes.Equal(content, expectedContent) {
			t.Fatalf("unexpected content for %s:\ngot:\n%s---\nexpected:\n%s---",
				file,
				string(content),
				string(expectedContent))
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
