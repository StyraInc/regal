package lsp

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/storage"

	"github.com/styrainc/regal/internal/lsp/cache"
	"github.com/styrainc/regal/internal/lsp/completions/refs"
	"github.com/styrainc/regal/internal/lsp/types"
	rparse "github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/hints"
	"github.com/styrainc/regal/pkg/linter"
	"github.com/styrainc/regal/pkg/report"
	"github.com/styrainc/regal/pkg/rules"
)

// diagnosticsRunOpts contains options for file and workspace linting.
type diagnosticsRunOpts struct {
	Cache            *cache.Cache
	RegalConfig      *config.Config
	WorkspaceRootURI string
	UpdateForRules   []string
	CustomRulesPath  string

	// File-specific
	FileURI string

	// Workspace-specific
	OverwriteAggregates bool
	AggregateReportOnly bool
}

// updateParse updates the module cache with the latest parse result for a given URI,
// if the module cannot be parsed, the parse errors are saved as diagnostics for the
// URI instead.
func updateParse(
	ctx context.Context,
	cache *cache.Cache,
	store storage.Store,
	fileURI string,
	builtins map[string]*ast.Builtin,
	version ast.RegoVersion,
) (bool, error) {
	content, ok := cache.GetFileContents(fileURI)
	if !ok {
		return false, fmt.Errorf("failed to get file contents for uri %q", fileURI)
	}

	lines := strings.Split(content, "\n")
	options := rparse.ParserOptions()
	options.RegoVersion = version

	module, err := rparse.ModuleWithOpts(fileURI, content, options)
	if err == nil {
		// if the parse was ok, clear the parse errors
		cache.SetParseErrors(fileURI, []types.Diagnostic{})
		cache.SetModule(fileURI, module)
		cache.SetSuccessfulParseLineCount(fileURI, len(lines))

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

	var astErrors []ast.Error

	// Check if err is of type ast.Errors
	var astErrs ast.Errors
	if errors.As(err, &astErrs) {
		for _, e := range astErrs {
			astErrors = append(astErrors, ast.Error{
				Code:     e.Code,
				Message:  e.Message,
				Location: e.Location,
			})
		}
	} else {
		// Check if err is a single ast.Error
		var singleAstErr *ast.Error
		if errors.As(err, &singleAstErr) {
			astErrors = append(astErrors, ast.Error{
				Code:     singleAstErr.Code,
				Message:  singleAstErr.Message,
				Location: singleAstErr.Location,
			})
		} else {
			// Unknown error type
			return false, fmt.Errorf("unknown error type: %T", err)
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

		//nolint:gosec
		diags = append(diags, types.Diagnostic{
			Severity: util.Pointer(uint(1)), // parse errors are the only error Diagnostic the server sends
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
			Source:  &key,
			Code:    strings.ReplaceAll(astError.Code, "_", "-"),
			CodeDescription: &types.CodeDescription{
				Href: link,
			},
		})
	}

	cache.SetParseErrors(fileURI, diags)

	if len(diags) == 0 {
		return false, errors.New("failed to parse module, but no errors were set as diagnostics")
	}

	return false, nil
}

func updateFileDiagnostics(ctx context.Context, opts diagnosticsRunOpts) error {
	if opts.OverwriteAggregates {
		return errors.New("OverwriteAggregates should not be set for updateFileDiagnostics")
	}

	if opts.AggregateReportOnly {
		return errors.New("AggregateReportOnly should not be set for updateFileDiagnostics")
	}

	module, ok := opts.Cache.GetModule(opts.FileURI)
	if !ok {
		// then there must have been a parse error
		return nil
	}

	contents, ok := opts.Cache.GetFileContents(opts.FileURI)
	if !ok {
		return fmt.Errorf("failed to get file contents for uri %q", opts.FileURI)
	}

	input := rules.NewInput(
		map[string]string{opts.FileURI: contents},
		map[string]*ast.Module{opts.FileURI: module},
	)

	regalInstance := linter.NewLinter().
		// needed to get the aggregateData for this file
		WithCollectQuery(true).
		// needed to get the aggregateData out so we can update the cache
		WithExportAggregates(true).
		WithInputModules(&input).
		WithPathPrefix(opts.WorkspaceRootURI)

	if opts.RegalConfig != nil {
		regalInstance = regalInstance.WithUserConfig(*opts.RegalConfig)
	}

	if opts.CustomRulesPath != "" {
		regalInstance = regalInstance.WithCustomRules([]string{opts.CustomRulesPath})
	}

	rpt, err := regalInstance.Lint(ctx)
	if err != nil {
		return fmt.Errorf("failed to lint: %w", err)
	}

	fileDiags := convertReportToDiagnostics(&rpt, opts.WorkspaceRootURI)

	for uri := range opts.Cache.GetAllFiles() {
		// if a file has parse errors, continue to show these until they're addressed
		parseErrs, ok := opts.Cache.GetParseErrors(uri)
		if ok && len(parseErrs) > 0 {
			continue
		}

		// For updateFileDiagnostics, we only update the file in question.
		if uri == opts.FileURI {
			fd, ok := fileDiags[uri]
			if !ok {
				fd = []types.Diagnostic{}
			}

			opts.Cache.SetFileDiagnosticsForRules(uri, opts.UpdateForRules, fd)
		}
	}

	opts.Cache.SetFileAggregates(opts.FileURI, rpt.Aggregates)

	return nil
}

func updateWorkspaceDiagnostics(ctx context.Context, opts diagnosticsRunOpts) error {
	if opts.FileURI != "" {
		return errors.New("FileURI should not be set for updateAllDiagnostics")
	}

	var err error

	modules := opts.Cache.GetAllModules()
	files := opts.Cache.GetAllFiles()

	regalInstance := linter.NewLinter().
		WithPathPrefix(opts.WorkspaceRootURI).
		// aggregates need only be exported if they're to be used to overwrite.
		WithExportAggregates(opts.OverwriteAggregates)

	if opts.RegalConfig != nil {
		regalInstance = regalInstance.WithUserConfig(*opts.RegalConfig)
	}

	if opts.CustomRulesPath != "" {
		regalInstance = regalInstance.WithCustomRules([]string{opts.CustomRulesPath})
	}

	if opts.AggregateReportOnly {
		regalInstance = regalInstance.WithAggregates(opts.Cache.GetFileAggregates())
	} else {
		input := rules.NewInput(files, modules)
		regalInstance = regalInstance.WithInputModules(&input)
	}

	rpt, err := regalInstance.Lint(ctx)
	if err != nil {
		return fmt.Errorf("failed to lint: %w", err)
	}

	fileDiags := convertReportToDiagnostics(&rpt, opts.WorkspaceRootURI)

	for uri := range files {
		parseErrs, ok := opts.Cache.GetParseErrors(uri)
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
		if opts.AggregateReportOnly {
			opts.Cache.SetFileDiagnosticsForRules(uri, opts.UpdateForRules, fd)
		} else {
			opts.Cache.SetFileDiagnostics(uri, fd)
		}
	}

	if opts.OverwriteAggregates {
		// clear all aggregates, and use these ones
		opts.Cache.SetAggregates(rpt.Aggregates)
	}

	return nil
}

func convertReportToDiagnostics(rpt *report.Report, workspaceRootURI string) map[string][]types.Diagnostic {
	fileDiags := make(map[string][]types.Diagnostic, len(rpt.Violations))

	// rangeValCopy necessary, as value copied in loop anyway
	//nolint:gocritic
	for _, item := range rpt.Violations {
		// here errors are presented as warnings, and warnings as info
		// to differentiate from parse errors
		severity := uint(2)
		if item.Level == "warning" {
			severity = 3
		}

		file := cmp.Or(item.Location.File, workspaceRootURI)
		source := "regal/" + item.Category

		fileDiags[file] = append(fileDiags[file], types.Diagnostic{
			Severity: &severity,
			Range:    getRangeForViolation(item),
			Message:  item.Description,
			Source:   &source,
			Code:     item.Title,
			CodeDescription: &types.CodeDescription{
				Href: fmt.Sprintf(
					"https://docs.styra.com/regal/rules/%s/%s",
					item.Category,
					item.Title,
				),
			},
		})
	}

	return fileDiags
}

//nolint:gosec
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

	return types.Range{Start: start, End: end}
}
