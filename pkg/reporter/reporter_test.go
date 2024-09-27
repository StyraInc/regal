package reporter

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/styrainc/regal/pkg/report"
)

func ptr(s string) *string {
	return &s
}

//nolint:gochecknoglobals
var rep = report.Report{
	Summary: report.Summary{
		FilesScanned:  3,
		NumViolations: 2,
		FilesFailed:   2,
		RulesSkipped:  1,
	},
	Violations: []report.Violation{
		{
			Title:       "breaking-the-law",
			Description: "Rego must not break the law!",
			Category:    "legal",
			Location: report.Location{
				File:   "a.rego",
				Row:    1,
				Column: 1,
				Text:   ptr("package illegal"),
				End: &report.Position{
					Row:    1,
					Column: 14,
				},
			},
			RelatedResources: []report.RelatedResource{
				{
					Description: "documentation",
					Reference:   "https://example.com/illegal",
				},
			},
			Level: "error",
		},
		{
			Title:       "questionable-decision",
			Description: "Questionable decision found",
			Category:    "really?",
			Location: report.Location{
				File:   "b.rego",
				Row:    22,
				Column: 18,
				Text:   ptr("default allow = true"),
			},
			RelatedResources: []report.RelatedResource{
				{
					Description: "documentation",
					Reference:   "https://example.com/questionable",
				},
			},
			Level: "warning",
		},
	},
	Notices: []report.Notice{
		{
			Title:       "rule-made-obsolete",
			Description: "Rule made obsolete by capability foo",
			Category:    "some-category",
			Severity:    "none",
			Level:       "notice",
		},
		{
			Title:       "rule-missing-capability",
			Description: "Rule missing capability bar",
			Category:    "some-category",
			Severity:    "warning",
			Level:       "notice",
		},
	},
}

func TestPrettyReporterPublish(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := NewPrettyReporter(&buf).Publish(context.Background(), rep); err != nil {
		t.Fatal(err)
	}

	expect := MustReadFile(t, "testdata/pretty/reporter.txt")
	if buf.String() != expect {
		t.Errorf("expected %s, got %s", expect, buf.String())
	}
}

func TestPrettyReporterPublishLongText(t *testing.T) {
	t.Parallel()

	longRep := report.Report{
		Summary: report.Summary{
			FilesScanned:  3,
			NumViolations: 1,
			FilesFailed:   0,
			RulesSkipped:  0,
		},
		Violations: []report.Violation{
			{
				Title:       "long-violation",
				Description: "violation with a long description",
				Category:    "long",
				Location: report.Location{
					File:   "b.rego",
					Row:    22,
					Column: 18,
					Text:   ptr(strings.Repeat("long,", 1000)),
				},
				RelatedResources: []report.RelatedResource{
					{
						Description: "documentation",
						Reference:   "https://example.com/to-long",
					},
				},
				Level: "warning",
			},
		},
	}

	var buf bytes.Buffer
	if err := NewPrettyReporter(&buf).Publish(context.Background(), longRep); err != nil {
		t.Fatal(err)
	}

	if expect := MustReadFile(t, "testdata/pretty/reporter-long-text.txt"); expect != buf.String() {
		t.Errorf("expected %s, got %s", expect, buf.String())
	}
}

func TestPrettyReporterPublishNoViolations(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := NewPrettyReporter(&buf).Publish(context.Background(), report.Report{}); err != nil {
		t.Fatal(err)
	}

	if buf.String() != "0 files linted. No violations found.\n" {
		t.Errorf(`expected "0 files linted. No violations found.\n", got %q`, buf.String())
	}
}

func TestCompactReporterPublish(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := NewCompactReporter(&buf).Publish(context.Background(), rep); err != nil {
		t.Fatal(err)
	}

	expect := `+--------------+------------------------------+
|   Location   |         Description          |
+--------------+------------------------------+
| a.rego:1:1   | Rego must not break the law! |
| b.rego:22:18 | Questionable decision found  |
+--------------+------------------------------+

`
	if buf.String() != expect {
		t.Errorf("expected %s, got %s", expect, buf.String())
	}
}

func TestCompactReporterPublishNoViolations(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := NewCompactReporter(&buf).Publish(context.Background(), report.Report{}); err != nil {
		t.Fatal(err)
	}

	if buf.String() != "\n" {
		t.Errorf("expected %q, got %q", "", buf.String())
	}
}

func TestJSONReporterPublish(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := NewJSONReporter(&buf).Publish(context.Background(), rep); err != nil {
		t.Fatal(err)
	}

	if expect := MustReadFile(t, "testdata/json/reporter.json"); expect != buf.String() {
		t.Errorf("expected %q, got %q", expect, buf.String())
	}
}

func TestJSONReporterPublishNoViolations(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := NewJSONReporter(&buf).Publish(context.Background(), report.Report{}); err != nil {
		t.Fatal(err)
	}

	if expect := MustReadFile(t, "testdata/json/reporter-no-violations.json"); expect != buf.String() {
		t.Errorf("expected %q, got %q", expect, buf.String())
	}
}

// nolint:paralleltest
func TestGitHubReporterPublish(t *testing.T) {
	// Can't use t.Parallel() here because t.Setenv() forbids that
	t.Setenv("GITHUB_STEP_SUMMARY", "")

	var buf bytes.Buffer
	if err := NewGitHubReporter(&buf).Publish(context.Background(), rep); err != nil {
		t.Fatal(err)
	}

	if expectTable := "Rule:         \tbreaking-the-law"; !strings.Contains(buf.String(), expectTable) {
		t.Errorf("expected table output %q, got %q", expectTable, buf.String())
	}

	//nolint:lll
	expectGithub := `::error file=a.rego,line=1,col=1::Rego must not break the law!. To learn more, see: https://example.com/illegal
::warning file=b.rego,line=22,col=18::Questionable decision found. To learn more, see: https://example.com/questionable
`
	if !strings.Contains(buf.String(), expectGithub) {
		t.Errorf("expected workflow command output %q, got %q", expectGithub, buf.String())
	}
}

// nolint:paralleltest
func TestGitHubReporterPublishNoViolations(t *testing.T) {
	// Can't use t.Parallel() here because t.Setenv() forbids that
	t.Setenv("GITHUB_STEP_SUMMARY", "")

	var buf bytes.Buffer
	if err := NewGitHubReporter(&buf).Publish(context.Background(), report.Report{}); err != nil {
		t.Fatal(err)
	}

	if buf.String() != "0 files linted. No violations found.\n" {
		t.Errorf("expected %q, got %q", "", buf.String())
	}
}

func TestSarifReporterPublish(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := NewSarifReporter(&buf).Publish(context.Background(), rep); err != nil {
		t.Fatal(err)
	}

	if expect := MustReadFile(t, "testdata/sarif/reporter.json"); buf.String() != expect {
		t.Errorf("expected %s, got %s", expect, buf.String())
	}
}

// https://github.com/StyraInc/regal/issues/514
func TestSarifReporterViolationWithoutRegion(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := NewSarifReporter(&buf).Publish(context.Background(), report.Report{
		Violations: []report.Violation{
			{
				Title:       "opa-fmt",
				Description: "File should be formatted with `opa fmt`",
				Category:    "style",
				Location: report.Location{
					File: "policy.rego",
				},
				RelatedResources: []report.RelatedResource{
					{
						Description: "documentation",
						Reference:   "https://docs.styra.com/regal/rules/style/opa-fmt",
					},
				},
				Level: "error",
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if expect := MustReadFile(t, "testdata/sarif/reporter-no-region.json"); buf.String() != expect {
		t.Errorf("expected %s, got %s", expect, buf.String())
	}
}

func TestSarifReporterPublishNoViolations(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := NewSarifReporter(&buf).Publish(context.Background(), report.Report{}); err != nil {
		t.Fatal(err)
	}

	if expect := MustReadFile(t, "testdata/sarif/reporter-no-violation.json"); buf.String() != expect {
		t.Errorf("expected %s, got %s", expect, buf.String())
	}
}

// nolint:lll // the expected output is unfortunately longer than the allowed max line length
func TestJUnitReporterPublish(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := NewJUnitReporter(&buf).Publish(context.Background(), rep); err != nil {
		t.Fatal(err)
	}

	if expect := MustReadFile(t, "testdata/junit/reporter.xml"); buf.String() != expect {
		t.Errorf("expected %s, got %s", expect, buf.String())
	}
}

func TestJUnitReporterPublishNoViolations(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := NewJUnitReporter(&buf).Publish(context.Background(), report.Report{}); err != nil {
		t.Fatal(err)
	}

	if expect := "<testsuites name=\"regal\"></testsuites>\n"; buf.String() != expect {
		t.Errorf("expected \n%s, got \n%s", expect, buf.String())
	}
}

func MustReadFile(t *testing.T, path string) string {
	t.Helper()

	bs, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(bs)
}
