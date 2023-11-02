package reporter

import (
	"bytes"
	"context"
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

	pr := NewPrettyReporter(&buf)

	err := pr.Publish(context.Background(), rep)
	if err != nil {
		t.Fatal(err)
	}

	// TODO(anders): I cannot for the life of me get this to work using a raw string 🫠
	expect := "Rule:         \tbreaking-the-law                \nDescription:  \tRego must not break the law!    \nCategory:     \tlegal                           \nLocation:     \ta.rego:1:1                      \nText:         \tpackage illegal                 \nDocumentation:\thttps://example.com/illegal     \n              \nRule:         \tquestionable-decision           \nDescription:  \tQuestionable decision found     \nCategory:     \treally?                         \nLocation:     \tb.rego:22:18                    \nText:         \tdefault allow = true            \nDocumentation:\thttps://example.com/questionable\n\n3 files linted. 2 violations found in 2 files. 1 rule skipped:\n- rule-missing-capability: Rule missing capability bar\n\n" //nolint: lll
	if buf.String() != expect {
		t.Errorf("expected %s, got %s", expect, buf.String())
	}
}

func TestPrettyReporterPublishNoViolations(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	pr := NewPrettyReporter(&buf)

	err := pr.Publish(context.Background(), report.Report{})
	if err != nil {
		t.Fatal(err)
	}

	if buf.String() != "0 files linted. No violations found.\n" {
		t.Errorf(`expected "0 files linted. No violations found.\n", got %q`, buf.String())
	}
}

func TestCompactReporterPublish(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	cr := NewCompactReporter(&buf)

	err := cr.Publish(context.Background(), rep)
	if err != nil {
		t.Fatal(err)
	}

	expect := `a.rego:1:1  	Rego must not break the law!
b.rego:22:18	Questionable decision found
`

	if buf.String() != expect {
		t.Errorf("expected %q, got %q", expect, buf.String())
	}
}

func TestCompactReporterPublishNoViolations(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	cr := NewCompactReporter(&buf)

	err := cr.Publish(context.Background(), report.Report{})
	if err != nil {
		t.Fatal(err)
	}

	if buf.String() != "\n" {
		t.Errorf("expected %q, got %q", "", buf.String())
	}
}

func TestJSONReporterPublish(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	jr := NewJSONReporter(&buf)

	err := jr.Publish(context.Background(), rep)
	if err != nil {
		t.Fatal(err)
	}

	expect := `{
  "violations": [
    {
      "title": "breaking-the-law",
      "description": "Rego must not break the law!",
      "category": "legal",
      "level": "error",
      "related_resources": [
        {
          "description": "documentation",
          "ref": "https://example.com/illegal"
        }
      ],
      "location": {
        "col": 1,
        "row": 1,
        "file": "a.rego",
        "text": "package illegal"
      }
    },
    {
      "title": "questionable-decision",
      "description": "Questionable decision found",
      "category": "really?",
      "level": "warning",
      "related_resources": [
        {
          "description": "documentation",
          "ref": "https://example.com/questionable"
        }
      ],
      "location": {
        "col": 18,
        "row": 22,
        "file": "b.rego",
        "text": "default allow = true"
      }
    }
  ],
  "notices": [
    {
      "title": "rule-made-obsolete",
      "description": "Rule made obsolete by capability foo",
      "category": "some-category",
      "level": "notice",
      "severity": "none"
    },
    {
      "title": "rule-missing-capability",
      "description": "Rule missing capability bar",
      "category": "some-category",
      "level": "notice",
      "severity": "warning"
    }
  ],
  "summary": {
    "files_scanned": 3,
    "files_failed": 2,
    "rules_skipped": 1,
    "num_violations": 2
  }
}
`
	if buf.String() != expect {
		t.Errorf("expected %q, got %q", expect, buf.String())
	}
}

func TestJSONReporterPublishNoViolations(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	jr := NewJSONReporter(&buf)

	err := jr.Publish(context.Background(), report.Report{})
	if err != nil {
		t.Fatal(err)
	}

	if buf.String() != `{
  "violations": [],
  "summary": {
    "files_scanned": 0,
    "files_failed": 0,
    "rules_skipped": 0,
    "num_violations": 0
  }
}
` {
		t.Errorf("expected %q, got %q", `{"violations":[]}`, buf.String())
	}
}

//nolint:paralleltest
func TestGitHubReporterPublish(t *testing.T) {
	// Can't use t.Parallel() here because t.Setenv() forbids that
	t.Setenv("GITHUB_STEP_SUMMARY", "")

	var buf bytes.Buffer

	cr := NewGitHubReporter(&buf)

	err := cr.Publish(context.Background(), rep)
	if err != nil {
		t.Fatal(err)
	}

	expectTable := "Rule:         \tbreaking-the-law"

	if !strings.Contains(buf.String(), expectTable) {
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

//nolint:paralleltest
func TestGitHubReporterPublishNoViolations(t *testing.T) {
	// Can't use t.Parallel() here because t.Setenv() forbids that
	t.Setenv("GITHUB_STEP_SUMMARY", "")

	var buf bytes.Buffer

	cr := NewGitHubReporter(&buf)

	err := cr.Publish(context.Background(), report.Report{})
	if err != nil {
		t.Fatal(err)
	}

	if buf.String() != "0 files linted. No violations found.\n" {
		t.Errorf("expected %q, got %q", "", buf.String())
	}
}
