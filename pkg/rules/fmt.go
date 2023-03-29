package rules

import (
	"bytes"
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/format"
	"github.com/styrainc/regal/pkg/report"
)

type OpaFmtRule struct{}

const (
	title                       = "opa-fmt"
	description                 = "File should be formatted with `opa fmt`"
	category                    = "style"
	relatedResourcesDescription = "documentation"
	relatedResourcesRef         = "https://docs.styra.com/regal/rules/opa-fmt"
)

func NewOpaFmtRule() *OpaFmtRule {
	return &OpaFmtRule{}
}

func (f *OpaFmtRule) Run(ctx context.Context, input Input) (*report.Report, error) {
	result := &report.Report{}

	for _, filename := range input.FileNames {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout when running %s rule: %w", title, ctx.Err())
		default:
			module := input.Modules[filename]
			unformatted := input.FileBytes[filename]

			formatted, err := format.Ast(module)
			if err != nil {
				return nil, fmt.Errorf("failed to format module %s: %w", filename, err)
			}

			if !bytes.Equal(formatted, unformatted) {
				violation := report.Violation{
					Title:       title,
					Description: description,
					Category:    category,
					RelatedResources: []report.RelatedResource{{
						Description: relatedResourcesDescription,
						Reference:   relatedResourcesRef,
					}},
					Location: report.Location{
						File: filename,
					},
				}
				result.Violations = append(result.Violations, violation)
			}
		}
	}

	return result, nil
}

func (f *OpaFmtRule) Name() string {
	return title
}

func (f *OpaFmtRule) Category() string {
	return category
}
