package metrics

import (
	"github.com/open-policy-agent/opa/profiler"

	"github.com/styrainc/regal/pkg/report"
)

const (
	RegalConfigSearch         = "regal_config_search"
	RegalConfigParse          = "regal_config_parse"
	RegalFilterIgnoredFiles   = "regal_filter_ignored_files"
	RegalFilterIgnoredModules = "regal_filter_ignored_modules"
	RegalInputParse           = "regal_input_parse"
	RegalLint                 = "regal_lint_total"
	RegalLintGo               = "regal_lint_go"
	RegalLintRego             = "regal_lint_rego"
	RegalLintRegoAggregate    = "regal_lint_rego_aggregate"
)

func FromExprStats(stats profiler.ExprStats) report.ProfileEntry {
	return report.ProfileEntry{
		Location:    stats.Location.String(),
		TotalTimeNs: stats.ExprTimeNs,
		NumEval:     stats.NumEval,
		NumRedo:     stats.NumRedo,
		NumGenExpr:  stats.NumGenExpr,
	}
}
