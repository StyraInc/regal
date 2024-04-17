package fixes

import (
	"testing"

	"github.com/open-policy-agent/opa/ast"
)

func TestUseAssignmentOperator(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		runtimeOptions *RuntimeOptions

		beforeFix []byte
		afterFix  []byte

		fixExpected bool
	}{
		"no change": {
			beforeFix: []byte(`package test

allow := true
`),
			afterFix: []byte(`package test

allow := true
`),
			fixExpected:    false,
			runtimeOptions: &RuntimeOptions{},
		},
		"no change because no location": {
			beforeFix: []byte(`package test

allow = true
`),
			afterFix: []byte(`package test

allow = true
`),
			fixExpected:    false,
			runtimeOptions: &RuntimeOptions{},
		},
		"single change": {
			beforeFix: []byte(`package test

allow = true
`),
			afterFix: []byte(`package test

allow := true
`),
			fixExpected: true,
			runtimeOptions: &RuntimeOptions{
				Locations: []ast.Location{
					{
						Row: 3,
						Col: 7,
					},
				},
			},
		},
		"bad change": {
			beforeFix: []byte(`package test

allow = true
`),
			afterFix: []byte(`package test

allow = true
`),
			fixExpected: false,
			runtimeOptions: &RuntimeOptions{
				Locations: []ast.Location{
					{
						Row: 1,
						Col: 1,
					},
				},
			},
		},
		"many changes": {
			beforeFix: []byte(`package test

allow = true if { u = 1 }

allow = true if { u = 2 }
`),
			afterFix: []byte(`package test

allow := true if { u = 1 }

allow := true if { u = 2 }
`),
			fixExpected: true,
			runtimeOptions: &RuntimeOptions{
				Locations: []ast.Location{
					{
						Row: 3,
						Col: 7,
					},
					{
						Row: 5,
						Col: 7,
					},
				},
			},
		},
		"different columns": {
			beforeFix: []byte(`package test

allow = true
 wow = true
  wowallow = true
`),
			afterFix: []byte(`package test

allow := true
 wow := true
  wowallow := true
`),
			fixExpected: true,
			runtimeOptions: &RuntimeOptions{
				Locations: []ast.Location{
					{
						Row: 3,
						Col: 7,
					},
					{
						Row: 4,
						Col: 6,
					},
					{
						Row: 5,
						Col: 12,
					},
				},
			},
		},
	}

	for testName, tc := range testCases {
		tc := tc

		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			uas := UseAssignmentOperator{}

			fixed, fixedContent, err := uas.Fix(tc.beforeFix, tc.runtimeOptions)
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
