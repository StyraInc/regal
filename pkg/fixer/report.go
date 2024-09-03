package fixer

import (
	"fmt"
	"slices"

	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/fixer/fixes"
)

// Report contains updated file contents and summary information about the fixes that were applied
// during a fix operation.
type Report struct {
	totalFixes uint
	fileFixes  map[string][]fixes.FixResult
	movedFiles map[string]string
}

func NewReport() *Report {
	return &Report{
		fileFixes:  make(map[string][]fixes.FixResult),
		movedFiles: make(map[string]string),
	}
}

func (r *Report) AddFileFix(file string, fix fixes.FixResult) {
	r.fileFixes[file] = append(r.fileFixes[file], fix)
	r.totalFixes++
}

func (r *Report) FixesForFile(file string) []fixes.FixResult {
	return r.fileFixes[file]
}

func (r *Report) MergeFixes(path1, path2 string) {
	r.fileFixes[path1] = append(r.FixesForFile(path1), r.FixesForFile(path2)...)
	delete(r.fileFixes, path2)
}

func (r *Report) RegisterOldPathForFile(newPath, oldPath string) error {
	if _, ok := r.movedFiles[newPath]; ok {
		return fmt.Errorf("file %s already moved from %s", newPath, r.movedFiles[newPath])
	}

	r.movedFiles[newPath] = oldPath

	return nil
}

func (r *Report) OldPathForFile(newPath string) (string, bool) {
	oldPath, ok := r.movedFiles[newPath]

	return oldPath, ok
}

func (r *Report) FixedFiles() []string {
	fixedFiles := util.Keys(r.fileFixes)

	// sort the files for deterministic output
	slices.Sort(fixedFiles)

	return fixedFiles
}

func (r *Report) TotalFixes() uint {
	// totalFixes is incremented for each unique violation that is fixed
	return r.totalFixes
}
