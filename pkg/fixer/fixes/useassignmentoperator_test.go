package fixes

import (
	"testing"

	"github.com/styrainc/regal/pkg/report"
)

func TestUseAssignmentOperator(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		contentAfterFix string
		fc              *FixCandidate
		fixExpected     bool
		runtimeOptions  *RuntimeOptions
	}{
		"no change": {
			fc: &FixCandidate{Filename: "test.rego", Contents: `package test

allow := true
`},
			contentAfterFix: `package test

allow := true
`,
			fixExpected:    false,
			runtimeOptions: &RuntimeOptions{},
		},
		"no change because no location": {
			fc: &FixCandidate{Filename: "test.rego", Contents: `package test

allow = true
`},
			contentAfterFix: `package test

allow = true
`,
			fixExpected:    false,
			runtimeOptions: &RuntimeOptions{},
		},
		"single change": {
			fc: &FixCandidate{Filename: "test.rego", Contents: `package test

allow = true
`},
			contentAfterFix: `package test

allow := true
`,
			fixExpected: true,
			runtimeOptions: &RuntimeOptions{
				Locations: []report.Location{
					{
						Row:    3,
						Column: 7,
					},
				},
			},
		},
		"bad change": {
			fc: &FixCandidate{
				Filename: "test.rego",
				Contents: `package test

allow = true
`,
			},
			contentAfterFix: `package test

allow = true
`,
			fixExpected: false,
			runtimeOptions: &RuntimeOptions{
				Locations: []report.Location{
					{
						Row:    1,
						Column: 1,
					},
				},
			},
		},
		"many changes": {
			fc: &FixCandidate{
				Filename: "test.rego",
				Contents: `package test

allow = true if { u = 1 }

allow = true if { u = 2 }
`,
			},
			contentAfterFix: `package test

allow := true if { u = 1 }

allow := true if { u = 2 }
`,
			fixExpected: true,
			runtimeOptions: &RuntimeOptions{
				Locations: []report.Location{
					{
						Row:    3,
						Column: 7,
					},
					{
						Row:    5,
						Column: 7,
					},
				},
			},
		},
		"different columns": {
			fc: &FixCandidate{
				Filename: "test.rego",
				Contents: `package test

allow = true
 wow = true
  wowallow = true
`,
			},
			contentAfterFix: `package test

allow := true
 wow := true
  wowallow := true
`,
			fixExpected: true,
			runtimeOptions: &RuntimeOptions{
				Locations: []report.Location{
					{
						Row:    3,
						Column: 7,
					},
					{
						Row:    4,
						Column: 6,
					},
					{
						Row:    5,
						Column: 12,
					},
				},
			},
		},
	}

	for testName, tc := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			uas := UseAssignmentOperator{}

			fixResults, err := uas.Fix(tc.fc, tc.runtimeOptions)
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

			if tc.fixExpected && fixedContent != tc.contentAfterFix {
				t.Fatalf(
					"unexpected content, got:\n%s---\nexpected:\n%s---",
					fixedContent,
					tc.contentAfterFix,
				)
			}
		})
	}
}
