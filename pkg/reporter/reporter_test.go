package reporter

import (
	"bytes"
	"testing"

	"github.com/styrainc/regal/pkg/report"
)

func ptr(s string) *string {
	return &s
}

//nolint:gochecknoglobals
var rep = report.Report{
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
		},
	},
}

func TestPrettyReporterPublish(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	pr := NewPrettyReporter(&buf)

	err := pr.Publish(rep)
	if err != nil {
		t.Fatal(err)
	}

	expect := `Found 2 violations in 2 files

Rule:         	breaking-the-law                
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
`

	if buf.String() != expect {
		t.Errorf("expected %q, got %q", expect, buf.String())
	}
}

func TestPrettyReporterPublishNoViolations(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	pr := NewPrettyReporter(&buf)

	err := pr.Publish(report.Report{})
	if err != nil {
		t.Fatal(err)
	}

	if buf.String() != "Found 0 violations in 0 files\n\n" {
		t.Errorf("expected %q, got %q", "Found 0 violations in 0 files\n\n", buf.String())
	}
}

func TestCompactReporterPublish(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	cr := NewCompactReporter(&buf)

	err := cr.Publish(rep)
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

	err := cr.Publish(report.Report{})
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

	err := jr.Publish(rep)
	if err != nil {
		t.Fatal(err)
	}

	expect := `{
  "violations": [
    {
      "title": "breaking-the-law",
      "description": "Rego must not break the law!",
      "category": "legal",
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
  ]
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

	err := jr.Publish(report.Report{})
	if err != nil {
		t.Fatal(err)
	}

	if buf.String() != `{
  "violations": []
}
` {
		t.Errorf("expected %q, got %q", `{"violations":[]}`, buf.String())
	}
}
