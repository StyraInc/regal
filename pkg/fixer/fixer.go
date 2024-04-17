package fixer

import (
	"context"
	"fmt"
	"slices"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"

	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/fixer/fp"
	"github.com/styrainc/regal/pkg/linter"
)

// Report contains updated file contents and summary information about the fixes that were applied
// during a fix operation.
type Report struct {
	totalFixes          int
	fileFixedViolations map[string]map[string]struct{}
}

func NewReport() *Report {
	return &Report{
		fileFixedViolations: make(map[string]map[string]struct{}),
	}
}

func (r *Report) SetFileFixedViolation(file string, violation string) {
	if _, ok := r.fileFixedViolations[file]; !ok {
		r.fileFixedViolations[file] = make(map[string]struct{})
	}

	_, ok := r.fileFixedViolations[file][violation]
	if !ok {
		r.fileFixedViolations[file][violation] = struct{}{}
		r.totalFixes++
	}
}

func (r *Report) FixedViolationsForFile(file string) []string {
	fixedViolations := make([]string, 0)
	for violation := range r.fileFixedViolations[file] {
		fixedViolations = append(fixedViolations, violation)
	}

	// sort the violations for deterministic output
	slices.Sort(fixedViolations)

	return fixedViolations
}

func (r *Report) FixedFiles() []string {
	fixedFiles := make([]string, 0)
	for file := range r.fileFixedViolations {
		fixedFiles = append(fixedFiles, file)
	}

	// sort the files for deterministic output
	slices.Sort(fixedFiles)

	return fixedFiles
}

func (r *Report) TotalFixes() int {
	// totalFixes is incremented for each unique violation that is fixed
	return r.totalFixes
}

// NewDefaultFixes returns a list of default fixes that are applied by the fix command.
// When a new fix is added, it should be added to this list.
func NewDefaultFixes() []fixes.Fix {
	return []fixes.Fix{
		&fixes.Fmt{},
		&fixes.Fmt{
			KeyOverride: "use-rego-v1",
			OPAFmtOpts: format.Opts{
				RegoVersion: ast.RegoV0CompatV1,
			},
		},
		&fixes.UseAssignmentOperator{},
		&fixes.NoWhitespaceComment{},
	}
}

type Fixer struct {
	registeredFixes map[string]any
}

func (f *Fixer) RegisterFixes(fixes ...fixes.Fix) {
	if f.registeredFixes == nil {
		f.registeredFixes = make(map[string]any)
	}

	for _, fix := range fixes {
		f.registeredFixes[fix.Key()] = fix
	}
}

func (f *Fixer) GetFixForKey(key string) (fixes.Fix, bool) {
	fix, ok := f.registeredFixes[key]
	if !ok {
		return nil, false
	}

	fixInstance, ok := fix.(fixes.Fix)
	if !ok {
		return nil, false
	}

	return fixInstance, true
}

func (f *Fixer) Fix(ctx context.Context, l *linter.Linter, fp fp.FileProvider) (*Report, error) {
	enabledRules, err := l.EnabledRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to determine enabled rules: %w", err)
	}

	var fixableEnabledRules []string

	for _, rule := range enabledRules {
		if _, ok := f.GetFixForKey(rule); ok {
			fixableEnabledRules = append(fixableEnabledRules, rule)
		}
	}

	fixReport := NewReport()

	for {
		fixMadeInIteration := false

		in, err := fp.ToInput()
		if err != nil {
			return nil, fmt.Errorf("failed to generate linter input: %w", err)
		}

		fixLinter := l.WithDisableAll(true).
			WithEnabledRules(fixableEnabledRules...).
			WithInputModules(&in)

		rep, err := fixLinter.Lint(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to lint before fixing: %w", err)
		}

		for _, violation := range rep.Violations {
			fixInstance, ok := f.GetFixForKey(violation.Title)
			if !ok {
				return nil, fmt.Errorf("no fix for violation %s", violation.Title)
			}

			fc, err := fp.GetFile(violation.Location.File)
			if err != nil {
				return nil, fmt.Errorf("failed to get file %s: %w", violation.Location.File, err)
			}

			fixCandidate := fixes.FixCandidate{
				Filename: violation.Location.File,
				Contents: fc,
			}

			fixResults, err := fixInstance.Fix(&fixCandidate, &fixes.RuntimeOptions{
				Locations: []ast.Location{
					{
						Row: violation.Location.Row,
						Col: violation.Location.Column,
					},
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to fix %s: %w", violation.Location.File, err)
			}

			if len(fixResults) > 0 {
				// Note: Only one content update fix result is currently supported
				fixResult := fixResults[0]

				err = fp.PutFile(violation.Location.File, fixResult.Contents)
				if err != nil {
					return nil, fmt.Errorf("failed to write fixed content to file %s: %w", violation.Location.File, err)
				}

				fixReport.SetFileFixedViolation(violation.Location.File, violation.Title)

				fixMadeInIteration = true
			}
		}

		if !fixMadeInIteration {
			break
		}
	}

	return fixReport, nil
}
