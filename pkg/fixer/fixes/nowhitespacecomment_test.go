package fixes

import (
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
)

func TestNoWhitespaceComment(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		contentAfterFix []byte
		fc              *FixCandidate
		fixExpected     bool
		runtimeOptions  *RuntimeOptions
	}{
		"no change needed": {
			fc: &FixCandidate{
				Filename: "test.rego",
				Contents: []byte(`package test\n

# this is a comment
`),
			},
			contentAfterFix: []byte(`package test\n

# this is a comment
`),
			fixExpected: false,
			runtimeOptions: &RuntimeOptions{
				Locations: []ast.Location{},
			},
		},
		"single change": {
			fc: &FixCandidate{
				Filename: "test.rego",
				Contents: []byte(`package test\n

#this is a comment
`),
			},
			contentAfterFix: []byte(`package test\n

# this is a comment
`),
			fixExpected: true,
			runtimeOptions: &RuntimeOptions{
				Locations: []ast.Location{
					{
						Row: 3,
						Col: 1,
					},
				},
			},
		},
		"many changes": {
			fc: &FixCandidate{
				Filename: "test.rego",
				Contents: []byte(`package test\n

#this is a comment
#this is a comment
#this is a comment
`),
			},
			contentAfterFix: []byte(`package test\n

# this is a comment
# this is a comment
# this is a comment
`),
			fixExpected: true,
			runtimeOptions: &RuntimeOptions{
				Locations: []ast.Location{
					{
						Row: 3,
						Col: 1,
					},
					{
						Row: 4,
						Col: 1,
					},
					{
						Row: 5,
						Col: 1,
					},
				},
			},
		},
		"many changes, different columns": {
			fc: &FixCandidate{
				Filename: "test.rego",
				Contents: []byte(`package test\n

#this is a comment
 #this is a comment
  #this is a comment
`),
			},
			contentAfterFix: []byte(`package test\n

# this is a comment
 # this is a comment
  # this is a comment
`),
			fixExpected: true,
			runtimeOptions: &RuntimeOptions{
				Locations: []ast.Location{
					{
						Row: 3,
						Col: 1,
					},
					{
						Row: 4,
						Col: 2,
					},
					{
						Row: 5,
						Col: 3,
					},
				},
			},
		},
	}

	for testName, tc := range testCases {
		nwc := NoWhitespaceComment{}

		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			fixResults, err := nwc.Fix(tc.fc, tc.runtimeOptions)
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

			if tc.fixExpected && string(fixedContent) != string(tc.contentAfterFix) {
				t.Fatalf("unexpected content, got:\n%s---\nexpected:\n%s---",
					string(fixedContent),
					string(tc.contentAfterFix))
			}
		})
	}
}
