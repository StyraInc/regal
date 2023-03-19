package linter

import (
	"context"
	"fmt"
	"os"
	"sort"

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
	ruleBundles  []*bundle.Bundle
	configBundle *bundle.Bundle
}

// NewLinter creates a new Regal linter.
func NewLinter() Linter {
	return Linter{}
}

// WithAddedBundle adds a bundle of rules and data to include in evaluation.
func (l Linter) WithAddedBundle(b bundle.Bundle) Linter {
	l.ruleBundles = append(l.ruleBundles, &b)

	return l
}

const regalUserConfig = "regal_user_config"

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

// Lint runs the linter on provided policies.
func (l Linter) Lint(ctx context.Context, result *loader.Result) (report.Report, error) {
	var regoArgs []func(*rego.Rego)
	regoArgs = append(regoArgs,
		rego.Query("report = data.regal.main.report"),
		rego.EnablePrintStatements(true),
		rego.PrintHook(topdown.NewPrintHook(os.Stderr)),
	)

	if l.configBundle != nil {
		regoArgs = append(regoArgs, rego.ParsedBundle(regalUserConfig, l.configBundle))
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
		return report.Report{}, err
	}

	// Maintain order across runs
	modules := util.Keys(result.Modules)
	sort.Strings(modules)

	aggregate := report.Report{}

	for _, name := range modules {
		resultSet, err := query.Eval(ctx, rego.EvalInput(result.Modules[name].Parsed))
		if err != nil {
			return report.Report{}, err
		}
		if len(resultSet) != 1 {
			return report.Report{}, fmt.Errorf("expected 1 item in resultset, got %d", len(resultSet))
		}

		r := report.Report{}
		if err = rio.JSONRoundTrip(resultSet[0].Bindings, &r); err != nil {
			return report.Report{}, err
		}

		aggregate.Violations = append(aggregate.Violations, r.Violations...)
	}

	return aggregate, nil
}
