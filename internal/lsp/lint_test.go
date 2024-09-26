package lsp

import (
	"reflect"
	"testing"

	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/pkg/report"
)

func TestConvertReportToDiagnostics(t *testing.T) {
	t.Parallel()

	violation1 := report.Violation{
		Level:       "error",
		Description: "Mock Error",
		Category:    "mock_category",
		Title:       "mock_title",
		Location:    report.Location{File: "file1"},
		IsAggregate: false,
	}
	violation2 := report.Violation{
		Level:       "warning",
		Description: "Mock Warning",
		Category:    "mock_category",
		Title:       "mock_title",
		Location:    report.Location{File: ""},
		IsAggregate: true,
	}

	rpt := &report.Report{
		Violations: []report.Violation{violation1, violation2},
	}

	expectedFileDiags := map[string][]types.Diagnostic{
		"file1": {
			{
				Severity: 2,
				Range:    getRangeForViolation(violation1),
				Message:  "Mock Error",
				Source:   "regal/mock_category",
				Code:     "mock_title",
				CodeDescription: &types.CodeDescription{
					Href: "https://docs.styra.com/regal/rules/mock_category/mock_title",
				},
			},
		},
		"workspaceRootURI": {
			{
				Severity: 3,
				Range:    getRangeForViolation(violation2),
				Message:  "Mock Warning",
				Source:   "regal/mock_category",
				Code:     "mock_title",
				CodeDescription: &types.CodeDescription{
					Href: "https://docs.styra.com/regal/rules/mock_category/mock_title",
				},
			},
		},
	}

	fileDiags := convertReportToDiagnostics(rpt, "workspaceRootURI")

	if !reflect.DeepEqual(fileDiags, expectedFileDiags) {
		t.Errorf("Expected file diagnostics: %v, got: %v", expectedFileDiags, fileDiags)
	}
}
