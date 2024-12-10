package commands

import (
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/lsp/types"
)

func TestParse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		ExecuteCommandParams types.ExecuteCommandParams
		ParseOptions         ParseOptions
		ExpectedTarget       string
		ExpectedLocation     *ast.Location
	}{
		"extract target only": {
			ExecuteCommandParams: types.ExecuteCommandParams{
				Command:   "example",
				Arguments: []interface{}{"target"},
			},
			ParseOptions:     ParseOptions{TargetArgIndex: 0},
			ExpectedTarget:   "target",
			ExpectedLocation: nil,
		},
		"extract target and location": {
			ExecuteCommandParams: types.ExecuteCommandParams{
				Command:   "example",
				Arguments: []interface{}{"target", "1", 2}, // different types for testing, but should be strings
			},
			ParseOptions:     ParseOptions{TargetArgIndex: 0, RowArgIndex: 1, ColArgIndex: 2},
			ExpectedTarget:   "target",
			ExpectedLocation: &ast.Location{Row: 1, Col: 2},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, err := Parse(tc.ExecuteCommandParams, tc.ParseOptions)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Target != tc.ExpectedTarget {
				t.Fatalf("expected target %q, got %q", tc.ExpectedTarget, result.Target)
			}

			if tc.ExpectedLocation == nil && result.Location != nil {
				t.Fatalf("expected location to be nil, got %v", result.Location)
			}

			if tc.ExpectedLocation != nil {
				if result.Location == nil {
					t.Fatalf("expected location to be %v, got nil", tc.ExpectedLocation)
				}

				if result.Location.Row != tc.ExpectedLocation.Row {
					t.Fatalf("expected row %d, got %d", tc.ExpectedLocation.Row, result.Location.Row)
				}

				if result.Location.Col != tc.ExpectedLocation.Col {
					t.Fatalf("expected col %d, got %d", tc.ExpectedLocation.Col, result.Location.Col)
				}
			}
		})
	}
}

func TestParse_Errors(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		ExecuteCommandParams types.ExecuteCommandParams
		ParseOptions         ParseOptions
		ExpectedError        string
	}{
		"error extracting target": {
			ExecuteCommandParams: types.ExecuteCommandParams{
				Command:   "example",
				Arguments: []interface{}{}, // empty and so nothing can be extracted
			},
			ParseOptions:  ParseOptions{TargetArgIndex: 0},
			ExpectedError: "no args supplied",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := Parse(tc.ExecuteCommandParams, tc.ParseOptions)
			if err == nil {
				t.Fatalf("expected error %q, got nil", tc.ExpectedError)
			}

			if err.Error() != tc.ExpectedError {
				t.Fatalf("expected error %q, got %q", tc.ExpectedError, err.Error())
			}
		})
	}
}
