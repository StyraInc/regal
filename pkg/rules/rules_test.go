package rules

import (
	"testing"

	"github.com/open-policy-agent/opa/ast"
)

func TestInputFromTextWithOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		Module      string
		RegoVersion ast.RegoVersion
	}{
		"regov1": {
			Module: `package test
p if { true }`,
			RegoVersion: ast.RegoV1,
		},
		"regov0": {
			Module: `package test
p { true }`,
			RegoVersion: ast.RegoV0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := InputFromTextWithOptions("p.rego", tc.Module, ast.ParserOptions{
				RegoVersion: tc.RegoVersion,
			})
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
