package fixer

import (
	"fmt"
	"io"
	"path/filepath"
	"slices"

	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/fixer/fixes"
)

// Reporter is responsible for outputting a fix report in a specific format.
type Reporter interface {
	Report(*Report) error
}

// ReporterForFormat returns a suitable Reporter for outputting a fix report in the given format.
func ReporterForFormat(format string, outputWriter io.Writer) (Reporter, error) {
	switch format {
	case "pretty":
		return NewPrettyReporter(outputWriter), nil
	default:
		return nil, fmt.Errorf("unsupported format %s", format)
	}
}

// PrettyReporter outputs a fix report in a human-readable format.
type PrettyReporter struct {
	outputWriter io.Writer
}

func NewPrettyReporter(outputWriter io.Writer) *PrettyReporter {
	return &PrettyReporter{
		outputWriter: outputWriter,
	}
}

func (r *PrettyReporter) Report(fixReport *Report) error {
	switch x := fixReport.TotalFixes(); x {
	case 0:
		fmt.Fprintln(r.outputWriter, "No fixes applied.")

		return nil
	case 1:
		fmt.Fprintln(r.outputWriter, "1 fix applied:")
	default:
		fmt.Fprintf(r.outputWriter, "%d fixes applied:\n", x)
	}

	byRoot := make(map[string]map[string][]fixes.FixResult)

	for file, fxs := range fixReport.fileFixes {
		for _, fix := range fxs {
			if _, ok := byRoot[fix.Root]; !ok {
				byRoot[fix.Root] = make(map[string][]fixes.FixResult)
			}

			byRoot[fix.Root][file] = append(byRoot[fix.Root][file], fix)
		}
	}

	i := 0

	movedNewLocs := util.MapInvert(fixReport.movedFiles)
	rootsSorted := util.Keys(byRoot)

	slices.Sort(rootsSorted)

	for _, root := range rootsSorted {
		if i > 0 {
			fmt.Fprintln(r.outputWriter)
		}

		fixesByFile := byRoot[root]
		files := util.Keys(fixesByFile)

		slices.Sort(files)
		fmt.Fprintf(r.outputWriter, "In project root: %s\n", root)

		for _, file := range files {
			fxs := fixesByFile[file]

			rel := relOrDefault(root, file, file)

			if _, ok := movedNewLocs[file]; !ok {
				if newLoc, ok := fixReport.movedFiles[file]; ok {
					fmt.Fprintf(r.outputWriter, "%s -> %s:\n", rel, relOrDefault(root, newLoc, newLoc))
				} else {
					fmt.Fprintf(r.outputWriter, "%s:\n", rel)
				}
			} else if len(fxs) == 1 {
				if oldLoc, ok := movedNewLocs[file]; ok {
					fmt.Fprintf(r.outputWriter, "%s -> %s:\n", relOrDefault(root, oldLoc, oldLoc), rel)
				}
			}

			for _, fix := range fxs {
				fmt.Fprintf(r.outputWriter, "- %s\n", fix.Title)
			}
		}

		i++
	}

	return nil
}

func relOrDefault(root, path, defaultValue string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return defaultValue
	}

	return rel
}
