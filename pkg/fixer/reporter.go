package fixer

import (
	"fmt"
	"io"
)

// Reporter writes a summary of the fixes applied to the output writer.
func Reporter(outputWriter io.Writer, fixReport *Report) {
	if fixReport.TotalFixes() == 0 {
		fmt.Fprintln(outputWriter, "No fixes applied.")

		return
	}

	if fixReport.TotalFixes() == 1 {
		fmt.Fprintln(outputWriter, "1 fix applied:")
	}

	if fixReport.TotalFixes() > 1 {
		fmt.Fprintf(outputWriter, "%d fixes applied:\n", fixReport.TotalFixes())
	}

	for _, file := range fixReport.FixedFiles() {
		fmt.Fprintf(outputWriter, "%s:\n", file)

		for _, f := range fixReport.FixedViolationsForFile(file) {
			fmt.Fprintf(outputWriter, "- %s\n", f)
		}
	}
}
