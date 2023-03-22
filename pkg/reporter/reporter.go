package reporter

import (
	"encoding/json"
	"fmt"
	"io"

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

type CompactReporter struct {
	out io.Writer
}

type JSONReporter struct {
	out io.Writer
}

func NewPrettyReporter(out io.Writer) PrettyReporter {
	return PrettyReporter{out: out}
}

func NewCompactReporter(out io.Writer) CompactReporter {
	return CompactReporter{out: out}
}

func NewJSONReporter(out io.Writer) JSONReporter {
	return JSONReporter{out: out}
}

func (tr PrettyReporter) Publish(r report.Report) error {
	table := uitable.New()
	table.MaxColWidth = 80
	table.Wrap = true
	numFiles := len(r.FileCount())

	plural := ""
	if numFiles == 0 || numFiles > 1 {
		plural = "s"
	}

	heading := fmt.Sprintf("Found %d violations in %d file%s\n\n", len(r.Violations), numFiles, plural)

	numViolations := len(r.Violations)

	for i, violation := range r.Violations {
		table.AddRow("Rule:", violation.Title)
		table.AddRow("Description:", violation.Description)
		table.AddRow("Category:", violation.Category)
		table.AddRow("Location:", violation.Location.String())

		if violation.Location.Text != nil {
			table.AddRow("Text:", string(violation.Location.Text))
		}

		table.AddRow("Documentation:", getDocumentationURL(violation))

		if i+1 < numViolations {
			table.AddRow("")
		}
	}

	end := ""
	if numViolations > 0 {
		end = "\n"
	}

	_, err := fmt.Fprint(tr.out, heading+table.String()+end)

	return err //nolint:wrapcheck
}

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
