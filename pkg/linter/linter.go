package linter

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/report"
)

// Linter stores data to use for linting.
type Linter struct {
	ruleBundles      []*bundle.Bundle
	configBundle     *bundle.Bundle
	customRulesPaths []string
}

const regalUserConfig = "regal_user_config"

// NewLinter creates a new Regal linter.
func NewLinter() Linter {
	return Linter{}
}

// WithAddedBundle adds a bundle of rules and data to include in evaluation.
func (l Linter) WithAddedBundle(b bundle.Bundle) Linter {
	l.ruleBundles = append(l.ruleBundles, &b)

	return l
}

// WithCustomRules adds custom rules for evaluation, from the Rego (and data) files provided at paths.
func (l Linter) WithCustomRules(paths []string) Linter {
	l.customRulesPaths = paths

	return l
}

// WithUserConfig provides config overrides set by the user.
func (l Linter) WithUserConfig(config map[string]any) Linter {
	l.configBundle = &bundle.Bundle{
		Manifest: bundle.Manifest{
			Roots:    &[]string{regalUserConfig},
			Metadata: map[string]any{"name": regalUserConfig},
		},
		Data: map[string]any{regalUserConfig: config},
	}

	return l
}

//nolint:gochecknoglobals
var query = ast.MustParseBody("report = data.regal.main.report")

// Lint runs the linter on provided policies.
func (l Linter) Lint(ctx context.Context, result *loader.Result) (report.Report, error) {
	var regoArgs []func(*rego.Rego)
	regoArgs = append(regoArgs,
		rego.ParsedQuery(query),
		rego.EnablePrintStatements(true),
		rego.PrintHook(topdown.NewPrintHook(os.Stderr)),
	)

	if l.configBundle != nil {
		regoArgs = append(regoArgs, rego.ParsedBundle(regalUserConfig, l.configBundle))
	}

	if l.customRulesPaths != nil {
		regoArgs = append(regoArgs, rego.Load(l.customRulesPaths, rio.ExcludeTestFilter()))
	}

	if l.ruleBundles != nil {
		for _, ruleBundle := range l.ruleBundles {
			var bundleName string
			if metadataName, ok := ruleBundle.Manifest.Metadata["name"].(string); ok {
				bundleName = metadataName
			}

			regoArgs = append(regoArgs, rego.ParsedBundle(bundleName, ruleBundle))
		}
	}

	query, err := rego.New(regoArgs...).PrepareForEval(ctx)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed preparing query for linting %w", err)
	}

	// Maintain order across runs
	modules := util.Keys(result.Modules)
	sort.Strings(modules)

	aggregate := report.Report{}

	for _, name := range modules {
		resultSet, err := query.Eval(ctx, rego.EvalInput(result.Modules[name].Parsed))
		if err != nil {
			return report.Report{}, fmt.Errorf("error encountered in query evaluation %w", err)
		}

		if len(resultSet) != 1 {
			//nolint:goerr113
			return report.Report{}, fmt.Errorf("expected 1 item in resultset, got %d", len(resultSet))
		}

		r := report.Report{}
		if err = rio.JSONRoundTrip(resultSet[0].Bindings, &r); err != nil {
			return report.Report{},
				fmt.Errorf("JSON rountrip failed for bindings: %v %w", resultSet[0].Bindings, err)
		}

		aggregate.Violations = append(aggregate.Violations, r.Violations...)
	}

	return aggregate, nil
}
