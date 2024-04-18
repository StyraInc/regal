package fixer

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/pkg/fixer/fileprovider"
	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/linter"
)

func NewFixer() *Fixer {
	return &Fixer{}
}

type Fixer struct {
	registeredFixes map[string]any
}

func (f *Fixer) RegisterFixes(fixes ...fixes.Fix) {
	if f.registeredFixes == nil {
		f.registeredFixes = make(map[string]any)
	}

	for _, fix := range fixes {
		f.registeredFixes[fix.Name()] = fix
	}
}

func (f *Fixer) GetFixForName(name string) (fixes.Fix, bool) {
	fix, ok := f.registeredFixes[name]
	if !ok {
		return nil, false
	}

	fixInstance, ok := fix.(fixes.Fix)
	if !ok {
		return nil, false
	}

	return fixInstance, true
}

func (f *Fixer) Fix(ctx context.Context, l *linter.Linter, fp fileprovider.FileProvider) (*Report, error) {
	enabledRules, err := l.DetermineEnabledRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to determine enabled rules: %w", err)
	}

	var fixableEnabledRules []string

	for _, rule := range enabledRules {
		if _, ok := f.GetFixForName(rule); ok {
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
			fixInstance, ok := f.GetFixForName(violation.Title)
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
