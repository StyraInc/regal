package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/open-policy-agent/opa/ast"

	rparse "github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/linter"
	"github.com/styrainc/regal/pkg/rules"
)

// updateParse updates the module cache with the latest parse result for a given URI,
// if the module cannot be parsed, the parse errors are saved as diagnostics for the
// URI instead.
func updateParse(cache *Cache, uri string) (bool, error) {
	content, ok := cache.GetFileContents(uri)
	if !ok {
		return false, fmt.Errorf("failed to get file contents for uri %q", uri)
	}

	module, err := rparse.Module(uri, content)
	if err != nil {
		unwrappedError := errors.Unwrap(err)

		jsonErrors, err := json.Marshal(unwrappedError)
		if err != nil {
			return false, fmt.Errorf("failed to marshal parse errors: %w", err)
		}

		var astErrors []astError
		if err := json.Unmarshal(jsonErrors, &astErrors); err != nil {
			return false, fmt.Errorf("failed to unmarshal parse errors: %w", err)
		}

		diags := make([]Diagnostic, 0)

		for _, astError := range astErrors {
			itemLen := 1

			line := astError.Location.Row - 1
			if line < 0 {
				line = 0
			}

			char := astError.Location.Col - 1
			if char < 0 {
				char = 0
			}

			diags = append(diags, Diagnostic{
				Severity: 0, // parse errors are always errors
				Range: Range{
					Start: Position{
						Line:      uint(line),
						Character: uint(char),
					},
					End: Position{
						Line:      uint(line),
						Character: uint(char + itemLen),
					},
				},
				Message: astError.Message,
				Source:  "regal/parse",
				Code: DiagnosticCode{
					Value: astError.Code,
					// TODO(charlieegan3): link directly to a specific error using matching
					Target: "https://docs.styra.com/opa/category/rego-parse-error",
				},
			})
		}

		cache.SetParseErrors(uri, diags)

		return false, nil
	}

	// if the parse was ok, clear the parse errors
	cache.SetParseErrors(uri, []Diagnostic{})
	cache.SetModule(uri, module)

	return true, nil
}

func updateFileDiagnostics(ctx context.Context, cache *Cache, regalConfig *config.Config, uri string) error {
	module, ok := cache.GetModule(uri)
	if !ok {
		// then there must have been a parse error
		return nil
	}

	contents, ok := cache.GetFileContents(uri)
	if !ok {
		return fmt.Errorf("failed to get file contents for uri %q", uri)
	}

	input := rules.NewInput(map[string]string{uri: contents}, map[string]*ast.Module{uri: module})

	regalInstance := linter.NewLinter().WithInputModules(&input)

	if regalConfig != nil {
		regalInstance = regalInstance.WithUserConfig(*regalConfig)
	}

	rpt, err := regalInstance.Lint(ctx)
	if err != nil {
		return fmt.Errorf("failed to lint: %w", err)
	}

	diags := make([]Diagnostic, 0)

	for _, item := range rpt.Violations {
		itemLen := 0
		if item.Location.Text != nil {
			itemLen = len(*item.Location.Text)
		}

		line := item.Location.Row - 1
		if line < 0 {
			line = 0
		}

		char := item.Location.Column - 1
		if char < 0 {
			char = 0
		}

		// here errors are presented as warnings, and warnings as info
		// to differentiate from parse errors
		severity := uint(2)
		if item.Level == "warning" {
			severity = 3
		}

		diags = append(diags, Diagnostic{
			Severity: severity,
			Range: Range{
				Start: Position{
					Line:      uint(line),
					Character: uint(char),
				},
				End: Position{
					Line:      uint(line),
					Character: uint(char + itemLen + 1),
				},
			},
			Message: item.Description,
			Source:  "regal/" + item.Category,
			Code: DiagnosticCode{
				Value: item.Title,
				Target: fmt.Sprintf(
					"https://docs.styra.com/regal/rules/%s/%s",
					item.Category,
					item.Title,
				),
			},
		})
	}

	cache.SetFileDiagnostics(uri, diags)

	return nil
}

func updateAllDiagnostics(ctx context.Context, cache *Cache, regalConfig *config.Config, detachedURI string) error {
	modules := cache.GetAllModules()
	files := cache.GetAllFiles()

	input := rules.NewInput(files, modules)

	regalInstance := linter.NewLinter().WithInputModules(&input)

	if regalConfig != nil {
		regalInstance = regalInstance.WithUserConfig(*regalConfig)
	}

	rpt, err := regalInstance.Lint(ctx)
	if err != nil {
		return fmt.Errorf("failed to lint: %w", err)
	}

	aggDiags := make(map[string][]Diagnostic)
	fileDiags := make(map[string][]Diagnostic)

	for _, item := range rpt.Violations {
		itemLen := 0
		if item.Location.Text != nil {
			itemLen = len(*item.Location.Text)
		}

		line := item.Location.Row - 1
		if line < 0 {
			line = 0
		}

		char := item.Location.Column - 1
		if char < 0 {
			char = 0
		}

		// here errors are presented as warnings, and warnings as info
		// to differentiate from parse errors
		severity := uint(2)
		if item.Level == "warning" {
			severity = 3
		}

		diag := Diagnostic{
			Severity: severity,
			Range: Range{
				Start: Position{
					Line:      uint(line),
					Character: uint(char),
				},
				End: Position{
					Line:      uint(line),
					Character: uint(char + itemLen + 1),
				},
			},
			Message: item.Description,
			Source:  "regal/" + item.Category,
			Code: DiagnosticCode{
				Value: item.Title,
				Target: fmt.Sprintf(
					"https://docs.styra.com/regal/rules/%s/%s",
					item.Category,
					item.Title,
				),
			},
		}

		// TODO(charlieegan3): it'd be nice to be able to only run aggregate rules in some cases, but for now, we
		// can just run all rules each time.
		if item.IsAggregate {
			if item.Location.File == "" {
				aggDiags[detachedURI] = append(aggDiags[detachedURI], diag)
			} else {
				aggDiags[item.Location.File] = append(aggDiags[item.Location.File], diag)
			}
		} else {
			fileDiags[item.Location.File] = append(fileDiags[item.Location.File], diag)
		}
	}

	// this lint contains authoritative information about all files
	// all diagnostics are cleared and replaced with the new lint
	for uri := range files {
		// if a file has parse errors, then we continue to show these until they're addressed
		// as if there are lint results they must be based on an old, parsed version of the file
		parseErrs, ok := cache.GetParseErrors(uri)
		if ok && len(parseErrs) > 0 {
			continue
		}

		ad, ok := aggDiags[uri]
		if !ok {
			ad = []Diagnostic{}
		}

		cache.SetAggregateDiagnostics(uri, ad)

		fd, ok := fileDiags[uri]
		if !ok {
			fd = []Diagnostic{}
		}

		cache.SetFileDiagnostics(uri, fd)
	}

	// handle the diagnostics for the workspace, under the detachedURI
	ad, ok := aggDiags[detachedURI]
	if !ok {
		ad = []Diagnostic{}
	}

	cache.SetAggregateDiagnostics(detachedURI, ad)

	return nil
}

// astError is copied from OPA but drop details as I (charlieegan3) had issues unmarsalling the field.
type astError struct {
	Code     string        `json:"code"`
	Message  string        `json:"message"`
	Location *ast.Location `json:"location,omitempty"`
}
