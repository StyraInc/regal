package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/storage"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/completions/refs"
	"github.com/styrainc/regal/internal/lsp/types"
	rparse "github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/hints"
	"github.com/styrainc/regal/pkg/linter"
	"github.com/styrainc/regal/pkg/report"
	"github.com/styrainc/regal/pkg/rules"
)

// updateParse updates the module cache with the latest parse result for a given URI,
// if the module cannot be parsed, the parse errors are saved as diagnostics for the
// URI instead.
func updateParse(
	ctx context.Context,
	cache *cache.Cache,
	store storage.Store,
	fileURI string,
	builtins map[string]*ast.Builtin,
) (bool, error) {
	content, ok := cache.GetFileContents(fileURI)
	if !ok {
		return false, fmt.Errorf("failed to get file contents for uri %q", fileURI)
	}

	lines := strings.Split(content, "\n")

	module, err := rparse.Module(fileURI, content)
	if err == nil {
		// if the parse was ok, clear the parse errors
		cache.SetParseErrors(fileURI, []types.Diagnostic{})

		cache.SetModule(fileURI, module)

		if err := PutFileMod(ctx, store, fileURI, module); err != nil {
			return false, fmt.Errorf("failed to update rego store with parsed module: %w", err)
		}

		definedRefs := refs.DefinedInModule(module, builtins)

		cache.SetFileRefs(fileURI, definedRefs)

		// TODO: consider how we use and generate these to avoid needing to have in the cache and the store
		var ruleRefs []string

		for _, ref := range definedRefs {
			if ref.Kind == types.Package {
				continue
			}

			ruleRefs = append(ruleRefs, ref.Label)
		}

		if err = PutFileRefs(ctx, store, fileURI, ruleRefs); err != nil {
			return false, fmt.Errorf("failed to update rego store with defined refs: %w", err)
		}

		return true, nil
	}

	unwrappedError := errors.Unwrap(err)

	// if the module is empty, then the unwrapped error is a single parse ast.Error
	// otherwise it's a slice of ast.Error.
	// When a user has an empty file, we still want to show this single error.
	var astErrors []astError

	var parseError *ast.Error

	if errors.As(unwrappedError, &parseError) {
		astErrors = append(astErrors, astError{
			Code:     parseError.Code,
			Message:  parseError.Message,
			Location: parseError.Location,
		})
	} else {
		jsonErrors, err := json.Marshal(unwrappedError)
		if err != nil {
			return false, fmt.Errorf("failed to marshal parse errors: %w", err)
		}

		if err := json.Unmarshal(jsonErrors, &astErrors); err != nil {
			return false, fmt.Errorf("failed to unmarshal parse errors: %w", err)
		}
	}

	diags := make([]types.Diagnostic, 0)

	for _, astError := range astErrors {
		line := max(astError.Location.Row-1, 0)

		lineLength := 1
		if line < len(lines) {
			lineLength = len(lines[line])
		}

		key := "regal/parse"
		link := "https://docs.styra.com/opa/category/rego-parse-error"

		hints, _ := hints.GetForError(err)
		if len(hints) > 0 {
			// there should only be one hint, so take the first
			key = hints[0]
			link = "https://docs.styra.com/opa/errors/" + hints[0]
		}

		diags = append(diags, types.Diagnostic{
			Severity: 1, // parse errors are the only error Diagnostic the server sends
			Range: types.Range{
				Start: types.Position{
					Line: uint(line),
					// we always highlight the whole line for parse errors to make them more visible
					Character: 0,
				},
				End: types.Position{
					Line:      uint(line),
					Character: uint(lineLength),
				},
			},
			Message: astError.Message,
			Source:  key,
			Code:    strings.ReplaceAll(astError.Code, "_", "-"),
			CodeDescription: &types.CodeDescription{
				Href: link,
			},
		})
	}

	cache.SetParseErrors(fileURI, diags)

	return false, nil
}

func updateFileDiagnostics(
	ctx context.Context,
	cache *cache.Cache,
	regalConfig *config.Config,
	fileURI string,
	workspaceRootDir string,
	updateDiagnosticsForRules []string,
) error {
	module, ok := cache.GetModule(fileURI)
	if !ok {
		// then there must have been a parse error
		return nil
	}

	contents, ok := cache.GetFileContents(fileURI)
	if !ok {
		return fmt.Errorf("failed to get file contents for uri %q", fileURI)
	}

	input := rules.NewInput(map[string]string{fileURI: contents}, map[string]*ast.Module{fileURI: module})

	regalInstance := linter.NewLinter().
		// needed to get the aggregateData for this file
		WithCollectQuery(true).
		// needed to get the aggregateData out so we can update the cache
		WithExportAggregates(true).
		WithInputModules(&input).
		WithRootDir(workspaceRootDir)

	if regalConfig != nil {
		regalInstance = regalInstance.WithUserConfig(*regalConfig)
	}

	rpt, err := regalInstance.Lint(ctx)
	if err != nil {
		return fmt.Errorf("failed to lint: %w", err)
	}

	fileDiags := convertReportToDiagnostics(&rpt, workspaceRootDir)

	files := cache.GetAllFiles()

	for uri := range files {
		// if a file has parse errors, continue to show these until they're addressed
		parseErrs, ok := cache.GetParseErrors(uri)
		if ok && len(parseErrs) > 0 {
			continue
		}

		// For updateFileDiagnostics, we only update the file in question.
		if uri == fileURI {
			fd, ok := fileDiags[uri]
			if !ok {
				fd = []types.Diagnostic{}
			}

			cache.SetFileDiagnosticsForRules(uri, updateDiagnosticsForRules, fd)
		}
	}

	cache.SetFileAggregates(fileURI, rpt.Aggregates)

	return nil
}

func updateAllDiagnostics(
	ctx context.Context,
	cache *cache.Cache,
	regalConfig *config.Config,
	workspaceRootDir string,
	overwriteAggregates bool,
	aggregatesReportOnly bool,
	updateDiagnosticsForRules []string,
) error {
	var err error

	modules := cache.GetAllModules()
	files := cache.GetAllFiles()

	regalInstance := linter.NewLinter().
		WithRootDir(workspaceRootDir).
		// aggregates need only be exported if they're to be used to overwrite.
		WithExportAggregates(overwriteAggregates)

	if regalConfig != nil {
		regalInstance = regalInstance.WithUserConfig(*regalConfig)
	}

	if aggregatesReportOnly {
		regalInstance = regalInstance.
			WithAggregates(cache.GetFileAggregates())
	} else {
		input := rules.NewInput(files, modules)
		regalInstance = regalInstance.WithInputModules(&input)
	}

	rpt, err := regalInstance.Lint(ctx)
	if err != nil {
		return fmt.Errorf("failed to lint: %w", err)
	}

	fileDiags := convertReportToDiagnostics(&rpt, workspaceRootDir)

	for uri := range files {
		parseErrs, ok := cache.GetParseErrors(uri)
		if ok && len(parseErrs) > 0 {
			continue
		}

		fd, ok := fileDiags[uri]
		if !ok {
			fd = []types.Diagnostic{}
		}

		// when only an aggregate report was run, then we must make sure to
		// only update diagnostics from these rules. So the report is
		// authoratative, but for those rules only.
		if aggregatesReportOnly {
			cache.SetFileDiagnosticsForRules(uri, updateDiagnosticsForRules, fd)
		} else {
			cache.SetFileDiagnostics(uri, fd)
		}
	}

	if overwriteAggregates {
		// clear all aggregates, and use these ones
		cache.SetAggregates(rpt.Aggregates)
	}

	return nil
}

func convertReportToDiagnostics(
	rpt *report.Report,
	workspaceRootURI string,
) map[string][]types.Diagnostic {
	fileDiags := make(map[string][]types.Diagnostic)

	for _, item := range rpt.Violations {
		// here errors are presented as warnings, and warnings as info
		// to differentiate from parse errors
		severity := uint(2)
		if item.Level == "warning" {
			severity = 3
		}

		diag := types.Diagnostic{
			Severity: severity,
			Range:    getRangeForViolation(item),
			Message:  item.Description,
			Source:   "regal/" + item.Category,
			Code:     item.Title,
			CodeDescription: &types.CodeDescription{
				Href: fmt.Sprintf(
					"https://docs.styra.com/regal/rules/%s/%s",
					item.Category,
					item.Title,
				),
			},
		}

		if item.Location.File == "" {
			fileDiags[workspaceRootURI] = append(fileDiags[workspaceRootURI], diag)
		} else {
			fileDiags[item.Location.File] = append(fileDiags[item.Location.File], diag)
		}
	}

	return fileDiags
}

// astError is copied from OPA but drop details as I (charlieegan3) had issues unmarshalling the field.
type astError struct {
	Location *ast.Location `json:"location,omitempty"`
	Code     string        `json:"code"`
	Message  string        `json:"message"`
}

func getRangeForViolation(item report.Violation) types.Range {
	start := types.Position{
		Line:      uint(max(item.Location.Row-1, 0)),
		Character: uint(max(item.Location.Column-1, 0)),
	}

	end := types.Position{}
	if item.Location.End != nil {
		end.Line = uint(max(item.Location.End.Row-1, 0))
		end.Character = uint(max(item.Location.End.Column-1, 0))
	} else {
		itemLen := 0
		if item.Location.Text != nil {
			itemLen = len(*item.Location.Text)
		}

		end.Line = start.Line
		end.Character = start.Character + uint(itemLen)
	}

	return types.Range{
		Start: start,
		End:   end,
	}
}
