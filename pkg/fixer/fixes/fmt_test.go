package fixes

import (
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"
)

func TestFmt(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fmt             *Fmt
		contentAfterFix []byte
		fc              *FixCandidate
		fixExpected     bool
	}{
		"no change": {
			fc:              &FixCandidate{Filename: "test.rego", Contents: []byte("package testutil\n")},
			contentAfterFix: []byte("package testutil\n"),
			fixExpected:     false,
			fmt:             &Fmt{},
		},
		"add a new line": {
			fc:              &FixCandidate{Filename: "test.rego", Contents: []byte("package testutil")},
			contentAfterFix: []byte("package testutil\n"),
			fmt:             &Fmt{},
			fixExpected:     true,
		},
		"add a new line before rule": {
			fc:              &FixCandidate{Filename: "test.rego", Contents: []byte("package testutil\nallow := true")},
			contentAfterFix: []byte("package testutil\n\nallow := true\n"),
			fmt:             &Fmt{},
			fixExpected:     true,
		},
		"rego v1": {
			fc: &FixCandidate{Filename: "test.rego", Contents: []byte("package testutil\nallow := true")},
			contentAfterFix: []byte(`package testutil

import rego.v1

allow := true
`),
			fmt: &Fmt{
				OPAFmtOpts: format.Opts{
					RegoVersion: ast.RegoV0CompatV1,
				},
			},
			fixExpected: true,
		},
	}

	for testName, tc := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			fixResults, err := tc.fmt.Fix(tc.fc, &RuntimeOptions{
				BaseDir: "",
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tc.fixExpected && len(fixResults) != 0 {
				t.Fatalf("unexpected fix applied")
			}

			if !tc.fixExpected {
				return
			}

			fixedContent := fixResults[0].Contents

			if string(fixedContent) != string(tc.contentAfterFix) {
				t.Fatalf("unexpected content, got:\n%s---\nexpected:\n%s---",
					string(fixedContent),
					string(tc.contentAfterFix))
			}
		})
	}
}
