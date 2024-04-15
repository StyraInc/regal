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

func TestFixer(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		rep             *report.Report
		fileContents    map[string]io.Reader
		opts            fixes.Options
		expectedUpdates map[string][]byte
	}{
		"opa-fmt only, single file": {
			rep: &report.Report{
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
			},
			fileContents: map[string]io.Reader{
				"main.rego": bytes.NewReader([]byte("package test\nimport rego.v1\nallow := true")),
			},
			opts: fixes.Options{
				Fmt: &fixes.FmtOptions{},
			},
			expectedUpdates: map[string][]byte{
				"main.rego": []byte(`package test

import rego.v1

allow := true
`),
			},
		},
		"opa-fmt and rego-v1, single file": {
			rep: &report.Report{
				Violations: []report.Violation{
					{
						Title:    "opa-fmt",
						Category: "style",
						Location: report.Location{
							File:   "main.rego",
							Row:    1,
							Column: 1,
						},
					},
					{
						Title:    "use-rego-v1",
						Category: "imports",
						Location: report.Location{
							File:   "main.rego",
							Row:    1,
							Column: 1,
						},
					},
				},
			},
			fileContents: map[string]io.Reader{
				"main.rego": bytes.NewReader([]byte("package test\nallow := true")),
			},
			opts: fixes.Options{
				Fmt: &fixes.FmtOptions{
					OPAFmtOpts: format.Opts{
						RegoVersion: ast.RegoV0CompatV1,
					},
				},
			},
			expectedUpdates: map[string][]byte{
				"main.rego": []byte(`package test

import rego.v1

allow := true
`),
			},
		},
	}

	for testName, tc := range testCases {
		tc := tc

		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			updates, err := Fix(tc.rep, tc.fileContents, tc.opts)
			if err != nil {
				t.Fatalf("failed to fix: %v", err)
			}

			if got, exp := len(updates), len(tc.expectedUpdates); got != exp {
				t.Fatalf("expected %d updates, got %d", exp, got)
			}

			for file, content := range updates {
				expectedContent, ok := tc.expectedUpdates[file]
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
		})
	}
}
