package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/gosuri/uitable"

	"github.com/styrainc/regal/pkg/report"
)

// Reporter releases linter reports in a format decided by the implementation.
type Reporter interface {
	// Publish releases a report to any appropriate target
	Publish(report.Report) error
}

// PrettyReporter is a Reporter for representing reports as tables.
type PrettyReporter struct {
	out io.Writer
}

// CompactReporter reports violations in a compact table.
type CompactReporter struct {
	out io.Writer
}

// JSONReporter reports violations as JSON.
type JSONReporter struct {
	out io.Writer
}

// NewPrettyReporter creates a new PrettyReporter.
func NewPrettyReporter(out io.Writer) PrettyReporter {
	return PrettyReporter{out: out}
}

// NewCompactReporter creates a new CompactReporter.
func NewCompactReporter(out io.Writer) CompactReporter {
	return CompactReporter{out: out}
}

// NewJSONReporter creates a new JSONReporter.
func NewJSONReporter(out io.Writer) JSONReporter {
	return JSONReporter{out: out}
}

// Publish prints a pretty report to the configured output.
func (tr PrettyReporter) Publish(r report.Report) error {
	table := buildPrettyViolationsTable(r.Violations)

	pluralScanned := ""
	if r.Summary.FilesScanned == 0 || r.Summary.FilesScanned > 1 {
		pluralScanned = "s"
	}

	footer := fmt.Sprintf("%d file%s linted.", r.Summary.FilesScanned, pluralScanned)

	if r.Summary.NumViolations == 0 { //nolint:nestif
		footer += " No violations found"
	} else {
		pluralViolations := ""
		if r.Summary.NumViolations > 1 {
			pluralViolations = "s"
		}

		footer += fmt.Sprintf(" %d violation%s found", r.Summary.NumViolations, pluralViolations)

		if r.Summary.FilesScanned > 1 && r.Summary.FilesFailed > 0 {
			pluralFailed := ""
			if r.Summary.FilesFailed > 1 {
				pluralFailed = "s"
			}

			footer += fmt.Sprintf(" in %d file%s", r.Summary.FilesFailed, pluralFailed)
		}
	}

	_, err := fmt.Fprint(tr.out, table+footer+".\n")

	return err //nolint:wrapcheck
}

func buildPrettyViolationsTable(violations []report.Violation) string {
	table := uitable.New()
	table.MaxColWidth = 120
	table.Wrap = false

	numViolations := len(violations)

	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	for i, violation := range violations {
		description := red(violation.Description)
		if violation.Level == "warning" {
			description = yellow(violation.Description)
		}

		table.AddRow(yellow("Rule:"), violation.Title)
		table.AddRow(yellow("Description:"), description)
		table.AddRow(yellow("Category:"), violation.Category)
		table.AddRow(yellow("Location:"), cyan(violation.Location.String()))

		if violation.Location.Text != nil {
			table.AddRow(yellow("Text:"), strings.TrimSpace(*violation.Location.Text))
		}

		table.AddRow(yellow("Documentation:"), cyan(getDocumentationURL(violation)))

		if i+1 < numViolations {
			table.AddRow("")
		}
	}

	end := ""
	if numViolations > 0 {
		end = "\n\n"
	}

	return table.String() + end
}

// Publish prints a compact report to the configured output.
func (tr CompactReporter) Publish(r report.Report) error {
	table := uitable.New()
	table.MaxColWidth = 80
	table.Wrap = true

	for _, violation := range r.Violations {
		table.AddRow(violation.Location.String(), violation.Description)
	}

	_, err := fmt.Fprintln(tr.out, table)

	return err //nolint:wrapcheck
}

// Publish prints a JSON report to the configured output.
func (tr JSONReporter) Publish(r report.Report) error {
	if r.Violations == nil {
		r.Violations = []report.Violation{}
	}

	bs, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshalling of report failed: %w", err)
	}

	_, err = fmt.Fprintln(tr.out, string(bs))

	return err //nolint:wrapcheck
}

func getDocumentationURL(violation report.Violation) string {
	for _, resource := range violation.RelatedResources {
		if resource.Description == "documentation" {
			return resource.Reference
		}
	}

	return ""
}
