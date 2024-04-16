package fixes

import (
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"
)

func TestFmt(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fmt            *Fmt
		runtimeOptions *RuntimeOptions

		beforeFix []byte
		afterFix  []byte

		fixExpected bool
	}{
		"no change": {
			beforeFix:   []byte("package testutil\n"),
			afterFix:    []byte("package testutil\n"),
			fixExpected: false,
			fmt:         &Fmt{},
			runtimeOptions: &RuntimeOptions{
				Metadata: RuntimeMetadata{
					Filename: "test.rego",
				},
			},
		},
		"add a new line": {
			beforeFix: []byte("package testutil"),
			afterFix:  []byte("package testutil\n"),
			fmt:       &Fmt{},
			runtimeOptions: &RuntimeOptions{
				Metadata: RuntimeMetadata{
					Filename: "test.rego",
				},
			},
			fixExpected: true,
		},
		"add a new line before rule": {
			beforeFix: []byte("package testutil\nallow := true"),
			afterFix:  []byte("package testutil\n\nallow := true\n"),
			fmt:       &Fmt{},
			runtimeOptions: &RuntimeOptions{
				Metadata: RuntimeMetadata{
					Filename: "test.rego",
				},
			},
			fixExpected: true,
		},
		"rego v1": {
			beforeFix: []byte("package testutil\nallow := true"),
			afterFix: []byte(`package testutil

import rego.v1

allow := true
`),
			fmt: &Fmt{
				OPAFmtOpts: format.Opts{
					RegoVersion: ast.RegoV0CompatV1,
				},
			},
			runtimeOptions: &RuntimeOptions{
				Metadata: RuntimeMetadata{
					Filename: "test.rego",
				},
			},
			fixExpected: true,
		},
	}

	for testName, tc := range testCases {
		tc := tc

		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			fixed, fixedContent, err := tc.fmt.Fix(tc.beforeFix, tc.runtimeOptions)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.fixExpected && !fixed {
				t.Fatalf("expected fix to be applied")
			}

			if string(fixedContent) != string(tc.afterFix) {
				t.Fatalf("unexpected content, got:\n%s---\nexpected:\n%s---",
					string(fixedContent),
					string(tc.afterFix))
			}
		})
	}
}
