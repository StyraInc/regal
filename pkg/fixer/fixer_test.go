package fixer

import (
	"bytes"
	"context"
	"slices"
	"testing"

	"github.com/styrainc/regal/pkg/fixer/fileprovider"
	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/linter"
)

func TestFixer(t *testing.T) {
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
	f.RegisterFixes(fixes.NewDefaultFixes()...)

	fixReport, err := f.Fix(context.Background(), &l, memfp)
	if err != nil {
		t.Fatalf("failed to fix: %v", err)
	}

	expectedFileFixedViolations := map[string][]string{
		// use-assignment-operator is not expected since use-rego-v1 also addresses this in this example
		"main.rego": {"no-whitespace-comment", "opa-fmt", "use-rego-v1"},
	}
	expectedFileContents := map[string][]byte{
		"main.rego": []byte(`package test

import rego.v1

allow := true

# no space

deny := true
`),
	}

	if got, exp := fixReport.TotalFixes(), 3; got != exp {
		t.Fatalf("expected %d fixed files, got %d", exp, got)
	}

	fpFiles, err := memfp.ListFiles()
	if err != nil {
		t.Fatalf("failed to list files: %v", err)
	}

	for _, file := range fpFiles {
		// check that the content is correct
		expectedContent, ok := expectedFileContents[file]
		if !ok {
			t.Fatalf("unexpected file %s", file)
		}

		content, err := memfp.GetFile(file)
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
		fixedViolations := fixReport.FixedViolationsForFile(file)

		expectedViolations, ok := expectedFileFixedViolations[file]
		if !ok {
			t.Fatalf("unexpected file waas fixed %s", file)
		}

		if !slices.Equal(fixedViolations, expectedViolations) {
			t.Fatalf("unexpected fixed violations for %s:\ngot: %v\nexpected: %v", file, fixedViolations, expectedViolations)
		}
	}
}
