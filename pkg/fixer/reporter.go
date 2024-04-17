package fixer

import (
	"fmt"
	"io"
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
	if fixReport.TotalFixes() == 0 {
		fmt.Fprintln(r.outputWriter, "No fixes applied.")

		return nil
	}

	if fixReport.TotalFixes() == 1 {
		fmt.Fprintln(r.outputWriter, "1 fix applied:")
	}

	if fixReport.TotalFixes() > 1 {
		fmt.Fprintf(r.outputWriter, "%d fixes applied:\n", fixReport.TotalFixes())
	}

	for _, file := range fixReport.FixedFiles() {
		fmt.Fprintf(r.outputWriter, "%s:\n", file)

		for _, f := range fixReport.FixedViolationsForFile(file) {
			fmt.Fprintf(r.outputWriter, "- %s\n", f)
		}
	}

	return nil
}
