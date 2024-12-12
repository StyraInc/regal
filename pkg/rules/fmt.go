package rules

import (
	"bytes"
	"context"
	"fmt"

	"github.com/anderseknert/roast/pkg/util"

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
			unformatted := util.StringToByteSlice(input.FileContent[filename])

			formatted, err := format.Ast(module)
			if err != nil {
				return nil, fmt.Errorf("failed to format module %s: %w", filename, err)
			}

			if !bytes.Equal(formatted, unformatted) {
				row := module.Package.Location.Row
				txt := module.Package.String()

				if lines := bytes.SplitN(unformatted, []byte("\n"), row+1); len(lines) > row {
					txt = util.ByteSliceToString(lines[row-1])
				}

				violation := report.Violation{
					Title:       title,
					Description: description,
					Category:    category,
					RelatedResources: []report.RelatedResource{{
						Description: relatedResourcesDescription,
						Reference:   f.Documentation(),
					}},
					Location: report.Location{
						File:   filename,
						Row:    row,
						Column: 1,
						Text:   &txt,
					},
					Level: f.ruleConfig.Level,
				}
				result.Violations = append(result.Violations, violation)
			}
		}
	}

	return result, nil
}

func (*OpaFmtRule) Name() string {
	return title
}

func (*OpaFmtRule) Category() string {
	return category
}

func (*OpaFmtRule) Description() string {
	return description
}

func (*OpaFmtRule) Documentation() string {
	return docs.CreateDocsURL(category, title)
}

func (f *OpaFmtRule) Config() config.Rule {
	return f.ruleConfig
}
