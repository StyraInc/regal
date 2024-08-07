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

	pr := NewPrettyReporter(&buf)

	err := pr.Publish(context.Background(), rep)
	if err != nil {
		t.Fatal(err)
	}

	// The actual output has trailing tabs, which go fmt strips out,
	// so we'll need to compare line by line with the trailing tabs removed
	expectLines := strings.Split(`Rule:         	breaking-the-law
Description:  	Rego must not break the law!
Category:     	legal
Location:     	a.rego:1:1
Text:         	package illegal
Documentation:	https://example.com/illegal

Rule:         	questionable-decision
Description:  	Questionable decision found
Category:     	really?
Location:     	b.rego:22:18
Text:         	default allow = true
Documentation:	https://example.com/questionable

3 files linted. 2 violations found in 2 files. 1 rule skipped:
- rule-missing-capability: Rule missing capability bar

`, "\n")

	for i, line := range strings.Split(buf.String(), "\n") {
		if strings.TrimRight(line, " \t") != expectLines[i] {
			t.Errorf("expected %q, got %q", expectLines[i], line)
		}
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
	pr := NewPrettyReporter(&buf)

	err := pr.Publish(context.Background(), longRep)
	if err != nil {
		t.Fatal(err)
	}

	//nolint:lll
	expectLines := strings.Split(`Rule:         	long-violation
Description:  	violation with a long description
Category:     	long
Location:     	b.rego:22:18
Text:         	long,long,long,long,long,long,long,long,long,long,long,long,long,long,long,long,long,long,long,long,long,long,long,lo...
Documentation:	https://example.com/to-long

3 files linted. 1 violation found.

`, "\n")

	for i, line := range strings.Split(buf.String(), "\n") {
		if got, want := strings.TrimSpace(line), expectLines[i]; got != want {
			t.Fatalf("expected\n%q\ngot\n%q", want, got)
		}
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
        "end": {
          "row": 1,
          "col": 14
        },
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

func TestSarifReporterPublish(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	sr := NewSarifReporter(&buf)

	err := sr.Publish(context.Background(), rep)
	if err != nil {
		t.Fatal(err)
	}

	expect := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "informationUri": "https://docs.styra.com/regal",
          "name": "Regal",
          "rules": [
            {
              "id": "breaking-the-law",
              "shortDescription": {
                "text": "Rego must not break the law!"
              },
              "helpUri": "https://example.com/illegal",
              "properties": {
                "category": "legal"
              }
            },
            {
              "id": "questionable-decision",
              "shortDescription": {
                "text": "Questionable decision found"
              },
              "helpUri": "https://example.com/questionable",
              "properties": {
                "category": "really?"
              }
            },
            {
              "id": "rule-missing-capability",
              "shortDescription": {
                "text": "Rule missing capability bar"
              },
              "properties": {
                "category": "some-category"
              }
            }
          ]
        }
      },
      "artifacts": [
        {
          "location": {
            "uri": "a.rego"
          },
          "length": -1
        },
        {
          "location": {
            "uri": "b.rego"
          },
          "length": -1
        }
      ],
      "results": [
        {
          "ruleId": "breaking-the-law",
          "ruleIndex": 0,
          "level": "error",
          "message": {
            "text": "Rego must not break the law!"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "a.rego"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 14
                }
              }
            }
          ]
        },
        {
          "ruleId": "questionable-decision",
          "ruleIndex": 1,
          "level": "warning",
          "message": {
            "text": "Questionable decision found"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "b.rego"
                },
                "region": {
                  "startLine": 22,
                  "startColumn": 18
                }
              }
            }
          ]
        },
        {
          "ruleId": "rule-missing-capability",
          "ruleIndex": 2,
          "kind": "informational",
          "level": "none",
          "message": {
            "text": "Rule missing capability bar"
          }
        }
      ]
    }
  ]
}`

	if buf.String() != expect {
		t.Errorf("expected %s, got %s", expect, buf.String())
	}
}

// https://github.com/StyraInc/regal/issues/514
func TestSarifReporterViolationWithoutRegion(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	sr := NewSarifReporter(&buf)

	err := sr.Publish(context.Background(), report.Report{
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
	})
	if err != nil {
		t.Fatal(err)
	}

	expect := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "informationUri": "https://docs.styra.com/regal",
          "name": "Regal",
          "rules": [
            {
              "id": "opa-fmt",
              "shortDescription": {
                "text": "File should be formatted with ` + "`opa fmt`" + `"
              },
              "helpUri": "https://docs.styra.com/regal/rules/style/opa-fmt",
              "properties": {
                "category": "style"
              }
            }
          ]
        }
      },
      "artifacts": [
        {
          "location": {
            "uri": "policy.rego"
          },
          "length": -1
        }
      ],
      "results": [
        {
          "ruleId": "opa-fmt",
          "ruleIndex": 0,
          "level": "error",
          "message": {
            "text": "File should be formatted with ` + "`opa fmt`" + `"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "policy.rego"
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`

	if buf.String() != expect {
		t.Errorf("expected %s, got %s", expect, buf.String())
	}
}

func TestSarifReporterPublishNoViolations(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	sr := NewSarifReporter(&buf)

	err := sr.Publish(context.Background(), report.Report{})
	if err != nil {
		t.Fatal(err)
	}

	expect := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "informationUri": "https://docs.styra.com/regal",
          "name": "Regal",
          "rules": []
        }
      },
      "results": []
    }
  ]
}`

	if buf.String() != expect {
		t.Errorf("expected %s, got %s", expect, buf.String())
	}
}

//nolint:lll // the expected output is unfortunately longer than the allowed max line length
func TestJUnitReporterPublish(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	sr := NewJUnitReporter(&buf)

	err := sr.Publish(context.Background(), rep)
	if err != nil {
		t.Fatal(err)
	}

	expect := `<testsuites name="regal" tests="2" failures="2">
	<testsuite name="a.rego" tests="1" failures="1" errors="0" id="0" time="">
		<testcase name="legal/breaking-the-law: Rego must not break the law!" classname="a.rego:1:1">
			<failure message="Rego must not break the law!. To learn more, see: https://example.com/illegal" type="error"><![CDATA[Rule: breaking-the-law
Description: Rego must not break the law!
Category: legal
Location: a.rego:1:1
Text: package illegal
Documentation: https://example.com/illegal]]></failure>
		</testcase>
	</testsuite>
	<testsuite name="b.rego" tests="1" failures="1" errors="0" id="0" time="">
		<testcase name="really?/questionable-decision: Questionable decision found" classname="b.rego:22:18">
			<failure message="Questionable decision found. To learn more, see: https://example.com/questionable" type="warning"><![CDATA[Rule: questionable-decision
Description: Questionable decision found
Category: really?
Location: b.rego:22:18
Text: default allow = true
Documentation: https://example.com/questionable]]></failure>
		</testcase>
	</testsuite>
</testsuites>
`

	if buf.String() != expect {
		t.Errorf("expected \n%s, got \n%s", expect, buf.String())
	}
}

func TestJUnitReporterPublishNoViolations(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	sr := NewJUnitReporter(&buf)

	err := sr.Publish(context.Background(), report.Report{})
	if err != nil {
		t.Fatal(err)
	}

	expect := `<testsuites name="regal"></testsuites>
`

	if buf.String() != expect {
		t.Errorf("expected \n%s, got \n%s", expect, buf.String())
	}
}
