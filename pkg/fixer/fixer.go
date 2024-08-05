package fixer

import (
	"bytes"
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/ast"

	"github.com/styrainc/regal/pkg/fixer/fileprovider"
	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/linter"
)

// NewFixer instantiates a Fixer.
func NewFixer() *Fixer {
	return &Fixer{
		registeredFixes:          make(map[string]any),
		registeredMandatoryFixes: make(map[string]any),
	}
}

// Fixer must be instantiated via NewFixer.
type Fixer struct {
	registeredFixes          map[string]any
	registeredMandatoryFixes map[string]any
}

// RegisterFixes sets the fixes that will be fixed if there are related linter
// violations that can be fixed by fixes.
func (f *Fixer) RegisterFixes(fixes ...fixes.Fix) {
	for _, fix := range fixes {
		if _, mandatory := f.registeredMandatoryFixes[fix.Name()]; !mandatory {
			f.registeredFixes[fix.Name()] = fix
		}
	}
}

// RegisterMandatoryFixes sets fixes which will be run before other registered
// fixes, against all files which are not ignored, regardless of linter
// violations.
func (f *Fixer) RegisterMandatoryFixes(fixes ...fixes.Fix) {
	for _, fix := range fixes {
		f.registeredMandatoryFixes[fix.Name()] = fix

		delete(f.registeredFixes, fix.Name())
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

func (f *Fixer) GetMandatoryFixForName(name string) (fixes.Fix, bool) {
	fix, ok := f.registeredMandatoryFixes[name]
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
	fixReport := NewReport()

	// first, run the mandatory fixes against all files
	for len(f.registeredMandatoryFixes) > 0 {
		fixMadeInIteration := false

		files, err := fp.ListFiles()
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		for _, file := range files {
			for fix := range f.registeredMandatoryFixes {
				fixInstance, ok := f.GetMandatoryFixForName(fix)
				if !ok {
					return nil, fmt.Errorf("no mandatory fix matched %s", fix)
				}

				fc, err := fp.GetFile(file)
				if err != nil {
					return nil, fmt.Errorf("failed to get file %s: %w", file, err)
				}

				fixCandidate := fixes.FixCandidate{Filename: file, Contents: fc}

				fixResults, err := fixInstance.Fix(&fixCandidate, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to fix %s: %w", file, err)
				}

				for _, fixResult := range fixResults {
					if !bytes.Equal(fc, fixResult.Contents) {
						err := fp.PutFile(file, fixResult.Contents)
						if err != nil {
							return nil, fmt.Errorf("failed to write fixed rego for file %s: %w", file, err)
						}

						fixReport.SetFileFixedViolation(file, fix)

						fixMadeInIteration = true
					}
				}
			}
		}

		if !fixMadeInIteration {
			break
		}
	}

	// if there are no registeredFixes (fixes that require a linter), then
	// we are done
	if len(f.registeredFixes) == 0 {
		return fixReport, nil
	}

	// next, run the fixes that require a linter violation trigger
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
