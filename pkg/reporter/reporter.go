//nolint:wrapcheck
package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/gosuri/uitable"

	"github.com/styrainc/regal/pkg/report"
)

// Reporter releases linter reports in a format decided by the implementation.
type Reporter interface {
	// Publish releases a report to any appropriate target
	Publish(context.Context, report.Report) error
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

// GitHubReporter reports violations in a format suitable for GitHub Actions.
type GitHubReporter struct {
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

// NewGitHubReporter creates a new GitHubReporter.
func NewGitHubReporter(out io.Writer) GitHubReporter {
	return GitHubReporter{out: out}
}

// Publish prints a pretty report to the configured output.
func (tr PrettyReporter) Publish(_ context.Context, r report.Report) error {
	table := buildPrettyViolationsTable(r.Violations)

	pluralScanned := ""
	if r.Summary.FilesScanned == 0 || r.Summary.FilesScanned > 1 {
		pluralScanned = "s"
	}

	footer := fmt.Sprintf("%d file%s linted.", r.Summary.FilesScanned, pluralScanned)

	if r.Summary.NumViolations == 0 { //nolint:nestif
		footer += " No violations found."
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

			footer += fmt.Sprintf(" in %d file%s.", r.Summary.FilesFailed, pluralFailed)
		} else {
			footer += "."
		}
	}

	if r.Summary.RulesSkipped > 0 {
		pluralSkipped := ""
		if r.Summary.RulesSkipped > 1 {
			pluralSkipped = "s"
		}

		footer += fmt.Sprintf(
			" %d rule%s skipped:\n",
			r.Summary.RulesSkipped,
			pluralSkipped,
		)

		for _, notice := range r.Notices {
			if notice.Severity != "none" {
				footer += fmt.Sprintf("- %s: %s\n", notice.Title, notice.Description)
			}
		}
	}

	_, err := fmt.Fprint(tr.out, table+footer+"\n")

	return err
}

func buildPrettyViolationsTable(violations []report.Violation) string {
	table := uitable.New()

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
func (tr CompactReporter) Publish(_ context.Context, r report.Report) error {
	table := uitable.New()
	table.MaxColWidth = 80
	table.Wrap = true

	for _, violation := range r.Violations {
		table.AddRow(violation.Location.String(), violation.Description)
	}

	_, err := fmt.Fprintln(tr.out, strings.TrimSuffix(table.String(), " "))

	return err
}

// Publish prints a JSON report to the configured output.
func (tr JSONReporter) Publish(_ context.Context, r report.Report) error {
	if r.Violations == nil {
		r.Violations = []report.Violation{}
	}

	bs, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshalling of report failed: %w", err)
	}

	_, err = fmt.Fprintln(tr.out, string(bs))

	return err
}

// Publish first prints the pretty formatted report to console for easy access in the logs. It then goes on
// to print the GitHub Actions annotations for each violation. Finally, it prints a summary of the report suitable
// for the GitHub Actions UI.
//
//nolint:nestif
func (tr GitHubReporter) Publish(ctx context.Context, r report.Report) error {
	err := NewPrettyReporter(tr.out).Publish(ctx, r)
	if err != nil {
		return err
	}

	if r.Violations == nil {
		r.Violations = []report.Violation{}
	}

	for _, violation := range r.Violations {
		_, err := fmt.Fprintf(tr.out,
			"::%s file=%s,line=%d,col=%d::%s\n",
			violation.Level,
			violation.Location.File,
			violation.Location.Row,
			violation.Location.Column,
			fmt.Sprintf("%s. To learn more, see: %s", violation.Description, getDocumentationURL(violation)),
		)
		if err != nil {
			return err
		}
	}

	pluralScanned := ""
	if r.Summary.FilesScanned == 0 || r.Summary.FilesScanned > 1 {
		pluralScanned = "s"
	}

	// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-job-summary
	if summaryFileLoc, ok := os.LookupEnv("GITHUB_STEP_SUMMARY"); ok && summaryFileLoc != "" {
		summaryFile, err := os.OpenFile(summaryFileLoc, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}

		defer func() {
			_ = summaryFile.Close()
		}()

		fmt.Fprintf(summaryFile, "### Regal Lint Report\n\n")

		fmt.Fprintf(summaryFile, "%d file%s linted.", r.Summary.FilesScanned, pluralScanned)

		if r.Summary.NumViolations == 0 { //nolint:nestif
			fmt.Fprintf(summaryFile, " No violations found")
		} else {
			pluralViolations := ""
			if r.Summary.NumViolations > 1 {
				pluralViolations = "s"
			}

			fmt.Fprintf(summaryFile, " %d violation%s found", r.Summary.NumViolations, pluralViolations)

			if r.Summary.FilesScanned > 1 && r.Summary.FilesFailed > 0 {
				pluralFailed := ""
				if r.Summary.FilesFailed > 1 {
					pluralFailed = "s"
				}

				fmt.Fprintf(summaryFile, " in %d file%s.", r.Summary.FilesFailed, pluralFailed)
				fmt.Fprintf(summaryFile, " See Files tab in PR for locations and details.\n\n")

				fmt.Fprintf(summaryFile, "#### Violations\n\n")

				for description, url := range getUniqueViolationURLs(r.Violations) {
					fmt.Fprintf(summaryFile, "* [%s](%s)\n", description, url)
				}
			}
		}
	}

	return nil
}

func getDocumentationURL(violation report.Violation) string {
	for _, resource := range violation.RelatedResources {
		if resource.Description == "documentation" {
			return resource.Reference
		}
	}

	return ""
}

func getUniqueViolationURLs(violations []report.Violation) map[string]string {
	urls := make(map[string]string)
	for _, violation := range violations {
		urls[violation.Description] = getDocumentationURL(violation)
	}

	return urls
}
