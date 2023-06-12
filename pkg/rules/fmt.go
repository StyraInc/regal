package rules

import (
	"bytes"
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/format"

	"github.com/styrainc/regal/internal/docs"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/report"
)

type OpaFmtRule struct {
	ruleConfig config.Rule
}

const (
	title                       = "opa-fmt"
	description                 = "File should be formatted with `opa fmt`"
	category                    = "style"
	relatedResourcesDescription = "documentation"
)

func NewOpaFmtRule(conf config.Config) *OpaFmtRule {
	ruleConf, ok := conf.Rules[category][title]
	if ok {
		return &OpaFmtRule{ruleConfig: ruleConf}
	}

	return &OpaFmtRule{ruleConfig: config.Rule{
		Level: "error",
	}}
}

func (f *OpaFmtRule) Run(ctx context.Context, input Input) (*report.Report, error) {
	result := &report.Report{}

	for _, filename := range input.FileNames {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout when running %s rule: %w", title, ctx.Err())
		default:
			module := input.Modules[filename]
			unformatted := []byte(input.FileContent[filename])

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
						Reference:   f.Documentation(),
					}},
					Location: report.Location{
						File: filename,
					},
					Level: f.ruleConfig.Level,
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

func (f *OpaFmtRule) Description() string {
	return description
}

func (f *OpaFmtRule) Documentation() string {
	return docs.DocsBaseURL + "/" + category + "/" + title + ".md"
}

func (f *OpaFmtRule) Config() config.Rule {
	return f.ruleConfig
}
