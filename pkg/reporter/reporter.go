//nolint:wrapcheck
package reporter

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/anderseknert/roast/pkg/encoding"
	rutil "github.com/anderseknert/roast/pkg/util"
	"github.com/fatih/color"
	"github.com/jstemmer/go-junit-report/v2/junit"
	"github.com/olekukonko/tablewriter"
	"github.com/owenrumney/go-sarif/v2/sarif"

	"github.com/styrainc/regal/internal/mode"
	"github.com/styrainc/regal/internal/novelty"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/fixer"
	"github.com/styrainc/regal/pkg/fixer/fixes"
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

// FestiveReporter reports violations in a format suitable for the holidays.
type FestiveReporter struct {
	out io.Writer
}

// SarifReporter reports violations in the SARIF (https://sarifweb.azurewebsites.net/) format.
type SarifReporter struct {
	out io.Writer
}

// JUnitReporter reports violations in the JUnit XML format
// (https://github.com/junit-team/junit5/blob/main/platform-tests/src/test/resources/jenkins-junit.xsd).
type JUnitReporter struct {
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

// NewFestiveReporter creates a new FestiveReporter.
func NewFestiveReporter(out io.Writer) FestiveReporter {
	return FestiveReporter{out: out}
}

// NewSarifReporter creates a new SarifReporter.
func NewSarifReporter(out io.Writer) SarifReporter {
	return SarifReporter{out: out}
}

// NewJUnitReporter creates a new JUnitReporter.
func NewJUnitReporter(out io.Writer) JUnitReporter {
	return JUnitReporter{out: out}
}

// Publish prints a pretty report to the configured output.
func (tr PrettyReporter) Publish(_ context.Context, r report.Report) error {
	table := buildPrettyViolationsTable(r.Violations)

	pluralScanned := ""
	if r.Summary.FilesScanned == 0 || r.Summary.FilesScanned > 1 {
		pluralScanned = "s"
	}

	footer := fmt.Sprintf("%d file%s linted.", r.Summary.FilesScanned, pluralScanned)

	if r.Summary.NumViolations == 0 {
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
	if err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	if !mode.Standalone { // don't bother advertising `regal fix` when not in standalone mode
		return nil
	}

	f := fixer.NewFixer()
	f.RegisterFixes(fixes.NewDefaultFixes()...)

	fixableViolations := make(map[string]struct{})
	fixableViolationsCount := 0

	for _, violation := range r.Violations {
		if fix, ok := f.GetFixForName(violation.Title); ok {
			if _, ok := fixableViolations[fix.Name()]; !ok {
				fixableViolations[fix.Name()] = struct{}{}
			}

			fixableViolationsCount++
		}
	}

	if fixableViolationsCount > 0 {
		violationKeys := util.Keys(fixableViolations)
		slices.Sort(violationKeys)

		_, err = fmt.Fprintf(
			tr.out,
			`
Hint: %d/%d violations can be automatically fixed (%s)
      Run regal fix --help for more details.
`,
			fixableViolationsCount,
			r.Summary.NumViolations,
			strings.Join(violationKeys, ", "),
		)
	}

	return err
}

// Publish prints a festive report to the configured output.
func (tr FestiveReporter) Publish(ctx context.Context, r report.Report) error {
	if os.Getenv("CI") == "" && len(r.Violations) == 0 {
		if err := novelty.HappyHolidays(); err != nil {
			return fmt.Errorf("novelty message display failed: %w", err)
		}
	}

	return NewPrettyReporter(tr.out).Publish(ctx, r)
}

func buildPrettyViolationsTable(violations []report.Violation) string {
	sb := &strings.Builder{}
	table := tablewriter.NewWriter(sb)

	table.SetNoWhiteSpace(true)
	table.SetTablePadding("\t")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetAutoWrapText(false)

	numViolations := len(violations)

	// Note: it's tempting to use table.SetColumnColor here, but for whatever reason, that requires using
	// table.SetHeader as well, and we don't want a header for this format.

	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	for i, violation := range violations {
		description := red(violation.Description)
		if violation.Level == "warning" {
			description = yellow(violation.Description)
		}

		table.Append([]string{yellow("Rule:"), violation.Title})

		// if there is no support for color, then we show the level in addition
		// so that the level of the violation is still clear
		if color.NoColor {
			table.Append([]string{"Level:", violation.Level})
		}

		table.Append([]string{yellow("Description:"), description})
		table.Append([]string{yellow("Category:"), violation.Category})
		// End location ignored here as it's not too interesting in this format and line:column
		// allows click-to-open.
		table.Append([]string{yellow("Location:"), cyan(violation.Location.String())})

		if violation.Location.Text != nil {
			if len(*violation.Location.Text) > 117 {
				table.Append([]string{yellow("Text:"), (*violation.Location.Text)[:117] + "..."})
			} else {
				table.Append([]string{yellow("Text:"), strings.TrimSpace(*violation.Location.Text)})
			}
		}

		table.Append([]string{yellow("Documentation:"), cyan(getDocumentationURL(violation))})

		if i+1 < numViolations {
			table.Append([]string{""})
		}
	}

	end := ""
	if numViolations > 0 {
		end = "\n"
	}

	table.Render()

	return sb.String() + end
}

// Publish prints a compact report to the configured output.
func (tr CompactReporter) Publish(_ context.Context, r report.Report) error {
	if len(r.Violations) == 0 {
		_, err := fmt.Fprintln(tr.out)

		return err
	}

	sb := &strings.Builder{}
	table := tablewriter.NewWriter(sb)

	table.SetHeader([]string{"Location", "Description"})
	table.SetAutoFormatHeaders(false)

	table.SetColWidth(80)
	table.SetAutoWrapText(true)

	for _, violation := range r.Violations {
		table.Append([]string{violation.Location.String(), violation.Description})
	}
	// plurals
	pluralScanned := ""
	if r.Summary.FilesScanned > 1 || r.Summary.FilesScanned == 0 {
		pluralScanned = "s"
	}

	pluralViolations := ""
	if r.Summary.NumViolations > 1 || r.Summary.NumViolations == 0 {
		pluralViolations = "s"
	}

	summary := fmt.Sprintf("%d file%s linted , %d violation%s found.",
		r.Summary.FilesScanned, pluralScanned,
		r.Summary.NumViolations, pluralViolations)
	// rendering the table
	table.Render()

	_, err := fmt.Fprintln(tr.out, strings.TrimSuffix(sb.String(), ""), summary)

	return err
}

// Publish prints a JSON report to the configured output.
func (tr JSONReporter) Publish(_ context.Context, r report.Report) error {
	if r.Violations == nil {
		r.Violations = []report.Violation{}
	}

	bs, err := encoding.JSON().MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshalling of report failed: %w", err)
	}

	_, err = fmt.Fprintln(tr.out, rutil.ByteSliceToString(bs))

	return err
}

// Publish first prints the pretty formatted report to console for easy access in the logs. It then goes on
// to print the GitHub Actions annotations for each violation. Finally, it prints a summary of the report suitable
// for the GitHub Actions UI.
func (tr GitHubReporter) Publish(ctx context.Context, r report.Report) error {
	if err := NewPrettyReporter(tr.out).Publish(ctx, r); err != nil {
		return err
	}

	if r.Violations == nil {
		r.Violations = []report.Violation{}
	}

	for _, violation := range r.Violations {
		if _, err := fmt.Fprintf(tr.out,
			"::%s file=%s,line=%d,col=%d::%s\n",
			violation.Level,
			violation.Location.File,
			violation.Location.Row,
			violation.Location.Column,
			fmt.Sprintf("%s. To learn more, see: %s", violation.Description, getDocumentationURL(violation)),
		); err != nil {
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

		if r.Summary.NumViolations == 0 {
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

// Publish prints a SARIF report to the configured output.
func (tr SarifReporter) Publish(_ context.Context, r report.Report) error {
	rep, err := sarif.New(sarif.Version210)
	if err != nil {
		return err
	}

	run := sarif.NewRunWithInformationURI("Regal", "https://docs.styra.com/regal")

	for _, violation := range r.Violations {
		pb := sarif.NewPropertyBag()
		pb.Add("category", violation.Category)

		run.AddRule(violation.Title).
			WithDescription(violation.Description).
			WithHelpURI(getDocumentationURL(violation)).
			WithProperties(pb.Properties)

		run.AddDistinctArtifact(violation.Location.File)

		run.CreateResultForRule(violation.Title).
			WithLevel(violation.Level).
			WithMessage(sarif.NewTextMessage(violation.Description)).
			AddLocation(getLocation(violation))
	}

	for _, notice := range r.Notices {
		if notice.Severity == "none" {
			// no need to report on notices like rules skipped due to
			// having been deprecated or made obsolete
			continue
		}

		pb := sarif.NewPropertyBag()
		pb.Add("category", notice.Category)

		run.AddRule(notice.Title).
			WithDescription(notice.Description).
			WithProperties(pb.Properties)

		run.CreateResultForRule(notice.Title).
			WithKind("informational").
			WithLevel("none").
			WithMessage(sarif.NewTextMessage(notice.Description))
	}

	rep.AddRun(run)

	return rep.PrettyWrite(tr.out)
}

func getLocation(violation report.Violation) *sarif.Location {
	physicalLocation := sarif.NewPhysicalLocation().
		WithArtifactLocation(
			sarif.NewSimpleArtifactLocation(violation.Location.File),
		)

	region := sarif.NewRegion().
		WithStartLine(violation.Location.Row).
		WithStartColumn(violation.Location.Column)

	if violation.Location.End != nil {
		region = region.WithEndLine(violation.Location.End.Row).WithEndColumn(violation.Location.End.Column)
	}

	if violation.Location.Row > 0 && violation.Location.Column > 0 {
		physicalLocation = physicalLocation.WithRegion(region)
	}

	return sarif.NewLocationWithPhysicalLocation(physicalLocation)
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

// Publish prints a JUnit XML report to the configured output.
func (tr JUnitReporter) Publish(_ context.Context, r report.Report) error {
	testSuites := junit.Testsuites{
		Name: "regal",
	}

	// group by file & sort by file
	files := make([]string, 0)
	violationsPerFile := map[string][]report.Violation{}

	for _, violation := range r.Violations {
		files = append(files, violation.Location.File)
		violationsPerFile[violation.Location.File] = append(violationsPerFile[violation.Location.File], violation)
	}

	sort.Strings(files)

	for _, file := range files {
		testsuite := junit.Testsuite{
			Name: file,
		}

		for _, violation := range violationsPerFile[file] {
			text := ""
			if violation.Location.Text != nil {
				text = strings.TrimSpace(*violation.Location.Text)
			}

			testsuite.AddTestcase(junit.Testcase{
				Name:      fmt.Sprintf("%s/%s: %s", violation.Category, violation.Title, violation.Description),
				Classname: violation.Location.String(),
				Failure: &junit.Result{
					Message: fmt.Sprintf("%s. To learn more, see: %s", violation.Description, getDocumentationURL(violation)),
					Type:    violation.Level,
					Data: fmt.Sprintf("Rule: %s\nDescription: %s\nCategory: %s\nLocation: %s\nText: %s\nDocumentation: %s",
						violation.Title,
						violation.Description,
						violation.Category,
						violation.Location.String(),
						text,
						getDocumentationURL(violation)),
				},
			})
		}

		testSuites.AddSuite(testsuite)
	}

	return testSuites.WriteXML(tr.out)
}
