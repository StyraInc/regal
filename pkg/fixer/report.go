package fixer

import "slices"

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
