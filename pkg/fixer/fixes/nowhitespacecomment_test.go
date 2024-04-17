package fixes

import (
	"testing"

	"github.com/open-policy-agent/opa/ast"
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
			fixExpected:    false,
			runtimeOptions: &RuntimeOptions{},
		},
		"no change made because no location": {
			fc: &FixCandidate{
				Filename: "test.rego",
				Contents: []byte(`package test\n

#this is a comment
`),
			},
			contentAfterFix: []byte(`package test\n

#this is a comment
`),
			fixExpected:    false,
			runtimeOptions: &RuntimeOptions{},
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
		"bad change": {
			fc: &FixCandidate{
				Filename: "test.rego",
				Contents: []byte(`package test\n

#this is a comment
`),
			},
			contentAfterFix: []byte(`package test\n

#this is a comment
`),
			fixExpected: false,
			runtimeOptions: &RuntimeOptions{
				Locations: []ast.Location{
					{
						Row: 3,
						Col: 9, // this is wrong and should not be fixed
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
		tc := tc

		nwc := NoWhitespaceComment{}

		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			fixed, fixedContent, err := nwc.Fix(tc.fc, tc.runtimeOptions)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.fixExpected != fixed {
				t.Fatalf("unexpected fixed value, got: %t, expected: %t", fixed, tc.fixExpected)
			}

			if tc.fixExpected && string(fixedContent) != string(tc.contentAfterFix) {
				t.Fatalf("unexpected content, got:\n%s---\nexpected:\n%s---",
					string(fixedContent),
					string(tc.contentAfterFix))
			}
		})
	}
}
