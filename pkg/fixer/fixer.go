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

	slices.Sort(fixedFiles)

	return fixedFiles
}

func (r *Report) FixedViolationsForFile(file string) []string {
	fixedViolations := make([]string, 0)
	for violation := range r.fileFixedViolations[file] {
		fixedViolations = append(fixedViolations, violation)
	}

	slices.Sort(fixedViolations)

	return fixedViolations
}

func (r *Report) TotalFixes() int {
	return r.totalFixes
}

type FixToggles struct {
	OPAFmt       bool
	OPAFmtRegoV1 bool
}

func (f *FixToggles) IsEnabled(key string) bool {
	if f == nil {
		return false
	}

	switch key {
	case "opa-fmt":
		return f.OPAFmt
	case "use-rego-v1":
		return f.OPAFmtRegoV1
	}

	return false
}

func NewDefaultFixes() []fixes.Fix {
	return []fixes.Fix{
		&fixes.Fmt{},
		&fixes.Fmt{
			OPAFmtOpts: format.Opts{
				RegoVersion: ast.RegoV0CompatV1,
			},
		},
		&fixes.UseAssignmentOperator{},
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

func (f *Fixer) Fix(rep *report.Report, readers map[string]io.Reader) (*Report, error) {
	filesToFix, err := computeFilesToFix(f, rep, readers)
	if err != nil {
		return nil, fmt.Errorf("failed to determine files to fix: %w", err)
	}

	fixReport := NewReport()

	for file, content := range filesToFix {
		for _, violation := range rep.Violations {
			fixInstance, ok := f.GetFixForKey(violation.Title)
			if !ok {
				continue
			}

			fixed, fixedContent, err := fixInstance.Fix(content, &fixes.RuntimeOptions{
				Metadata: fixes.RuntimeMetadata{
					Filename: file,
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

	return fixReport, nil
}

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

		// skip violations that are not enabled
		if _, ok := f.registeredFixes[violation.Title]; !ok {
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
