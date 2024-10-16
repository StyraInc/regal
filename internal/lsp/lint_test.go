package lsp

import (
	"context"
	"reflect"
	"testing"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/types"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/pkg/config"
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

func TestLintWithConfigIgnoreWildcards(t *testing.T) {
	t.Parallel()

	conf := &config.Config{
		Rules: map[string]config.Category{
			"idiomatic": {
				"directory-package-mismatch": config.Rule{
					Level: "ignore",
				},
			},
		},
	}

	contents := "package p\n\nimport rego.v1\n\ncamelCase := 1\n"
	module := parse.MustParseModule(contents)
	workspacePath := "/workspace"
	fileURI := "file:///workspace/ignore/p.rego"
	state := cache.NewCache()

	state.SetFileContents(fileURI, contents)
	state.SetModule(fileURI, module)
	state.SetFileDiagnostics(fileURI, []types.Diagnostic{})

	if err := updateFileDiagnostics(
		context.Background(),
		state,
		conf,
		fileURI,
		workspacePath,
		[]string{"prefer-snake-case"},
	); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	diagnostics, _ := state.GetFileDiagnostics(fileURI)
	if len(diagnostics) != 1 {
		t.Fatalf("Expected one diagnostic item, got %v", diagnostics)
	}

	if diagnostics[0].Code != "prefer-snake-case" {
		t.Errorf("Expected diagnostic code to be prefer-snake-case, got %v", diagnostics[0].Code)
	}

	// Clear the diagnostic and update the config with a wildcard ignore
	// for any file in the ignore directory.

	state.SetFileDiagnostics(fileURI, []types.Diagnostic{})

	conf.Rules["style"] = config.Category{
		"prefer-snake-case": config.Rule{
			Level: "error",
			Ignore: &config.Ignore{
				Files: []string{"ignore/**"},
			},
		},
	}

	if err := updateFileDiagnostics(
		context.Background(),
		state,
		conf,
		fileURI,
		workspacePath,
		[]string{"prefer-snake-case"},
	); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	diagnostics, _ = state.GetFileDiagnostics(fileURI)
	if len(diagnostics) != 0 {
		t.Fatalf("Expected no diagnostics, got %v", diagnostics)
	}
}
