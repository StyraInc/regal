package fixer

import (
	"fmt"
	"io"
	"slices"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"

	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/report"
)

// Report contains updated file contents and summary information about the fixes that were applied
// during a fix operation.
type Report struct {
	totalFixes          int
	fileFixedViolations map[string]map[string]struct{}
	fileContents        map[string][]byte
}

func NewReport() *Report {
	return &Report{
		fileFixedViolations: make(map[string]map[string]struct{}),
		fileContents:        make(map[string][]byte),
	}
}

func (r *Report) SetFileContents(file string, content []byte) {
	r.fileContents[file] = content
}

func (r *Report) GetFileContents(file string) ([]byte, bool) {
	content, ok := r.fileContents[file]

	return content, ok
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

func (r *Report) FileContents() map[string][]byte {
	return r.fileContents
}

func (r *Report) FixedFiles() []string {
	fixedFiles := make([]string, 0)
	for file := range r.fileContents {
		fixedFiles = append(fixedFiles, file)
	}

	// sort the files for deterministic output
	slices.Sort(fixedFiles)

	return fixedFiles
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

// OrderedFixes returns the fixes in the order they should be applied.
// Fixes that are marked as WholeFile are applied last since they
// can add new lines that would affect the fixes for line based violations.
func (f *Fixer) OrderedFixes() []fixes.Fix {
	orderedFixes := make([]fixes.Fix, 0)
	wholeFileFixes := make([]fixes.Fix, 0)

	for _, fix := range f.registeredFixes {
		fixInstance, ok := fix.(fixes.Fix)
		if !ok {
			continue
		}

		if fixInstance.WholeFile() {
			wholeFileFixes = append(wholeFileFixes, fixInstance)

			continue
		}

		orderedFixes = append(orderedFixes, fixInstance)
	}

	return append(orderedFixes, wholeFileFixes...)
}

func (f *Fixer) Fix(rep *report.Report, readers map[string]io.Reader) (*Report, error) {
	filesToFix, err := computeFilesToFix(f, rep, readers)
	if err != nil {
		return nil, fmt.Errorf("failed to determine files to fix: %w", err)
	}

	fixReport := NewReport()

	for _, fixInstance := range f.OrderedFixes() {
		// fix by line
		for file, content := range filesToFix {
			for _, violation := range rep.Violations {
				if violation.Title != fixInstance.Key() {
					continue
				}

				// if the file has been fixed, use the fixed content from other fixes
				if fixedContent, ok := fixReport.GetFileContents(file); ok {
					content = fixedContent
				}

				fixed, fixedContent, err := fixInstance.Fix(content, &fixes.RuntimeOptions{
					Metadata: fixes.RuntimeMetadata{
						Filename: file,
					},
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

				if fixed {
					fixReport.SetFileContents(file, fixedContent)
					fixReport.SetFileFixedViolation(file, violation.Title)
				}
			}
		}
	}

	return fixReport, nil
}

// computeFilesToFix determines which files need to be fixed based on the violations in the report.
func computeFilesToFix(
	f *Fixer,
	rep *report.Report,
	readers map[string]io.Reader,
) (map[string][]byte, error) {
	filesToFix := make(map[string][]byte)

	// determine which files need to be fixed
	for _, violation := range rep.Violations {
		file := violation.Location.File

		// skip files already marked for fixing
		if _, ok := filesToFix[file]; ok {
			continue
		}

		// skip violations that the fixer has no fix for
		if _, ok := f.GetFixForKey(violation.Title); !ok {
			continue
		}

		if _, ok := readers[file]; !ok {
			return nil, fmt.Errorf("no reader for fixable file %s", file)
		}

		bs, err := io.ReadAll(readers[file])
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		filesToFix[violation.Location.File] = bs
	}

	return filesToFix, nil
}
