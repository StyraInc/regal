package fixer

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/open-policy-agent/opa/v1/ast"

	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/fixer/fileprovider"
	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/linter"
	"github.com/styrainc/regal/pkg/report"
)

type OnConflictOperation string

const (
	OnConflictError  OnConflictOperation = "error"
	OnConflictRename OnConflictOperation = "rename"
)

// NewFixer instantiates a Fixer.
func NewFixer() *Fixer {
	return &Fixer{
		registeredFixes:          make(map[string]any),
		registeredMandatoryFixes: make(map[string]any),
		registeredRoots:          make([]string, 0),
		onConflictOperation:      OnConflictError,
	}
}

// Fixer must be instantiated via NewFixer.
type Fixer struct {
	registeredFixes          map[string]any
	registeredMandatoryFixes map[string]any
	onConflictOperation      OnConflictOperation
	registeredRoots          []string
	versionsMap              map[string]ast.RegoVersion
}

// SetOnConflictOperation sets the fixer's behavior when a conflict occurs.
func (f *Fixer) SetOnConflictOperation(operation OnConflictOperation) {
	f.onConflictOperation = operation
}

// SetRegoVersionsMap sets the mapping of path prefixes to versions for the
// fixer to use when creating input for fixer runs.
func (f *Fixer) SetRegoVersionsMap(versionsMap map[string]ast.RegoVersion) {
	f.versionsMap = versionsMap
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

// RegisterRoots sets the roots of the files that will be fixed.
// Certain fixes may require the nearest root of the file to be known,
// as fix operations could involve things like moving files, which
// will be moved relative to their nearest root.
func (f *Fixer) RegisterRoots(roots ...string) {
	f.registeredRoots = append(f.registeredRoots, roots...)
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

	// Early return if there are no registered mandatory fixes
	if len(f.registeredMandatoryFixes) == 0 && len(f.registeredFixes) == 0 {
		return fixReport, nil
	}

	// Apply mandatory fixes
	if err := f.applyMandatoryFixes(fp, fixReport); err != nil {
		return nil, err
	}

	// If there are no registered fixes that require a linter, return the report
	if len(f.registeredFixes) == 0 {
		return fixReport, nil
	}

	// Apply fixes that require linter violation triggers
	if err := f.applyLinterFixes(ctx, l, fp, fixReport); err != nil {
		return nil, err
	}

	return fixReport, nil
}

func (f *Fixer) FixViolations(
	violations []report.Violation,
	fp fileprovider.FileProvider,
	config *config.Config,
) (*Report, error) {
	fixReport := NewReport()

	startingFiles, err := fp.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	for _, violation := range violations {
		file := violation.Location.File

		fixInstance, ok := f.GetFixForName(violation.Title)
		if !ok {
			return nil, fmt.Errorf("no fix for violation %s", violation.Title)
		}

		fc, err := fp.Get(file)
		if err != nil {
			return nil, fmt.Errorf("failed to get file %s: %w", file, err)
		}

		fixCandidate := fixes.FixCandidate{
			Filename: file,
			Contents: fc,
		}

		abs, err := filepath.Abs(file)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %s: %w", file, err)
		}

		fixResults, err := fixInstance.Fix(&fixCandidate, &fixes.RuntimeOptions{
			BaseDir: util.FindClosestMatchingRoot(abs, f.registeredRoots),
			Config:  config,
			Locations: []ast.Location{
				{
					Row: violation.Location.Row,
					Col: violation.Location.Column,
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fix %s: %w", file, err)
		}

		if len(fixResults) == 0 {
			continue
		}

		fixResult := fixResults[0]

		if fixResult.Rename != nil {
			err = f.handleRename(fp, fixReport, startingFiles, fixResult)
			if err != nil {
				return nil, fmt.Errorf("failed to handle rename: %w", err)
			}
		}

		// Write the fixed content to the file
		if err := fp.Put(file, fixResult.Contents); err != nil {
			return nil, fmt.Errorf("failed to write fixed content to file %s: %w", file, err)
		}

		fixReport.AddFileFix(file, fixResult)
	}

	return fixReport, nil
}

// applyMandatoryFixes handles the application of mandatory fixes to all files.
func (f *Fixer) applyMandatoryFixes(fp fileprovider.FileProvider, fixReport *Report) error {
	if len(f.registeredMandatoryFixes) == 0 {
		return nil
	}

	for {
		fixMadeInIteration := false

		files, err := fp.List()
		if err != nil {
			return fmt.Errorf("failed to list files: %w", err)
		}

		for _, file := range files {
			for fix := range f.registeredMandatoryFixes {
				fixInstance, ok := f.GetMandatoryFixForName(fix)
				if !ok {
					return fmt.Errorf("no mandatory fix matched %s", fix)
				}

				fc, err := fp.Get(file)
				if err != nil {
					return fmt.Errorf("failed to get file %s: %w", file, err)
				}

				fixCandidate := fixes.FixCandidate{
					Filename: file,
					Contents: fc,
				}

				fixResults, err := fixInstance.Fix(&fixCandidate, &fixes.RuntimeOptions{
					BaseDir: util.FindClosestMatchingRoot(file, f.registeredRoots),
				})
				if err != nil {
					return fmt.Errorf("failed to fix %s: %w", file, err)
				}

				for _, fixResult := range fixResults {
					if fc != fixResult.Contents {
						if err := fp.Put(file, fixResult.Contents); err != nil {
							return fmt.Errorf("failed to write fixed rego for file %s: %w", file, err)
						}

						fixReport.AddFileFix(file, fixResult)

						fixMadeInIteration = true
					}
				}
			}
		}

		if !fixMadeInIteration {
			break
		}
	}

	return nil
}

// applyLinterFixes handles the application of fixes that require linter violation triggers.
func (f *Fixer) applyLinterFixes(
	ctx context.Context,
	l *linter.Linter,
	fp fileprovider.FileProvider,
	fixReport *Report,
) error {
	enabledRules, err := l.DetermineEnabledRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to determine enabled rules: %w", err)
	}

	var fixableEnabledRules []string

	for _, rule := range enabledRules {
		if _, ok := f.GetFixForName(rule); ok {
			fixableEnabledRules = append(fixableEnabledRules, rule)
		}
	}

	startingFiles, err := fp.List()
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if f.versionsMap == nil {
		return errors.New("rego versions map not set")
	}

	for {
		fixMadeInIteration := false

		in, err := fp.ToInput(f.versionsMap)
		if err != nil {
			return fmt.Errorf("failed to generate linter input: %w", err)
		}

		fixLinter := l.WithDisableAll(true).
			WithEnabledRules(fixableEnabledRules...).
			WithInputModules(&in)

		rep, err := fixLinter.Lint(ctx)
		if err != nil {
			return fmt.Errorf("failed to lint before fixing: %w", err)
		}

		if len(rep.Violations) == 0 {
			break
		}

		for _, violation := range rep.Violations {
			file := violation.Location.File

			fixInstance, ok := f.GetFixForName(violation.Title)
			if !ok {
				return fmt.Errorf("no fix for violation %s", violation.Title)
			}

			fc, err := fp.Get(file)
			if err != nil {
				return fmt.Errorf("failed to get file %s: %w", file, err)
			}

			fixCandidate := fixes.FixCandidate{
				Filename: file,
				Contents: fc,
			}

			config, err := l.GetConfig()
			if err != nil {
				return fmt.Errorf("failed to get config: %w", err)
			}

			abs, err := filepath.Abs(file)
			if err != nil {
				return fmt.Errorf("failed to get absolute path for %s: %w", file, err)
			}

			fixResults, err := fixInstance.Fix(&fixCandidate, &fixes.RuntimeOptions{
				BaseDir: util.FindClosestMatchingRoot(abs, f.registeredRoots),
				Config:  config,
				Locations: []ast.Location{
					{
						Row: violation.Location.Row,
						Col: violation.Location.Column,
					},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to fix %s: %w", file, err)
			}

			if len(fixResults) == 0 {
				continue
			}

			fixResult := fixResults[0]

			if fixResult.Rename != nil {
				err = f.handleRename(fp, fixReport, startingFiles, fixResult)
				if err != nil {
					return err
				}

				fixMadeInIteration = true

				break // Restart the loop after handling a rename
			}

			// Write the fixed content to the file
			if err := fp.Put(file, fixResult.Contents); err != nil {
				return fmt.Errorf("failed to write fixed content to file %s: %w", file, err)
			}

			fixReport.AddFileFix(file, fixResult)

			fixMadeInIteration = true
		}

		if !fixMadeInIteration {
			break
		}
	}

	return nil
}

// handleRename processes the rename operation and resolves conflicts if necessary.
func (f *Fixer) handleRename(
	fp fileprovider.FileProvider,
	fixReport *Report,
	startingFiles []string,
	fixResult fixes.FixResult,
) error {
	to := fixResult.Rename.ToPath
	from := fixResult.Rename.FromPath

	for {
		err := fp.Rename(from, to)
		if err == nil {
			// if there is no error, and no conflict, we have nothing to do
			break
		}

		var isConflict bool
		if errors.As(err, &fileprovider.RenameConflictError{}) {
			isConflict = true
		} else {
			return fmt.Errorf("failed to rename file: %w", err)
		}

		if isConflict {
			switch f.onConflictOperation {
			case OnConflictError:
				// OnConflictError is the default, these operations are taken to
				// ensure the correct state in the report for outputting the
				// verbose conflict report.
				// clean the old file to prevent repeated fixes
				if err := fp.Delete(from); err != nil {
					return fmt.Errorf("failed to delete file %s: %w", from, err)
				}

				if slices.Contains(startingFiles, to) {
					fixReport.RegisterConflictSourceFile(fixResult.Root, to, from)
				} else {
					fixReport.RegisterConflictManyToOne(fixResult.Root, to, from)
				}

				fixReport.AddFileFix(to, fixResult)
				fixReport.MergeFixes(to, from)
				fixReport.RegisterOldPathForFile(to, from)

				return nil
			case OnConflictRename:
				// OnConflictRename will select a new filename until there is no
				// conflict.
				to = renameCandidate(to)

				continue
			default:
				return fmt.Errorf("unsupported conflict operation: %v", f.onConflictOperation)
			}
		}
	}

	// update the fix result with the new path for consistency
	if to != fixResult.Rename.ToPath {
		fixResult.Rename.ToPath = to
	}

	fixReport.AddFileFix(to, fixResult)
	fixReport.MergeFixes(to, from)
	fixReport.RegisterOldPathForFile(to, from)

	return nil
}
