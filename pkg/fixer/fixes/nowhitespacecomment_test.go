package fixes

import (
	"testing"

	"github.com/open-policy-agent/opa/ast"
)

func TestNoWhitespaceComment(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		runtimeOptions *RuntimeOptions

		beforeFix []byte
		afterFix  []byte

		fixExpected bool
	}{
		"no change needed": {
			beforeFix: []byte(`package test\n

# this is a comment
`),
			afterFix: []byte(`package test\n

# this is a comment
`),
			fixExpected:    false,
			runtimeOptions: &RuntimeOptions{},
		},
		"no change made because no location": {
			beforeFix: []byte(`package test\n

#this is a comment
`),
			afterFix: []byte(`package test\n

#this is a comment
`),
			fixExpected:    false,
			runtimeOptions: &RuntimeOptions{},
		},
		"single change": {
			beforeFix: []byte(`package test\n

#this is a comment
`),
			afterFix: []byte(`package test\n

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
			beforeFix: []byte(`package test\n

#this is a comment
`),
			afterFix: []byte(`package test\n

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
			beforeFix: []byte(`package test\n

#this is a comment
#this is a comment
#this is a comment
`),
			afterFix: []byte(`package test\n

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
			beforeFix: []byte(`package test\n

#this is a comment
 #this is a comment
  #this is a comment
`),
			afterFix: []byte(`package test\n

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

			fixed, fixedContent, err := nwc.Fix(tc.beforeFix, tc.runtimeOptions)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.fixExpected != fixed {
				t.Fatalf("unexpected fixed value, got: %t, expected: %t", fixed, tc.fixExpected)
			}

			if tc.fixExpected && string(fixedContent) != string(tc.afterFix) {
				t.Fatalf("unexpected content, got:\n%s---\nexpected:\n%s---",
					string(fixedContent),
					string(tc.afterFix))
			}
		})
	}
}
