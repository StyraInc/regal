package linter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/imdario/mergo"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/builtins"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/report"
	"github.com/styrainc/regal/pkg/rules"
)

// Linter stores data to use for linting.
type Linter struct {
	ruleBundles      []*bundle.Bundle
	configBundle     *bundle.Bundle
	customRulesPaths []string
	combinedConfig   *config.Config
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

var query = ast.MustParseBody("violations = data.regal.main.report") //nolint:gochecknoglobals

// Lint runs the linter on provided policies.
func (l Linter) Lint(ctx context.Context, input rules.Input) (report.Report, error) {
	aggregate := report.Report{}

	goReport, err := l.lintWithGoRules(ctx, input)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to lint using Go rules: %w", err)
	}

	aggregate.Violations = append(aggregate.Violations, goReport.Violations...)

	regoReport, err := l.lintWithRegoRules(ctx, input)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to lint using Rego rules: %w", err)
	}

	aggregate.Violations = append(aggregate.Violations, regoReport.Violations...)

	return aggregate, nil
}

func (l Linter) lintWithGoRules(ctx context.Context, input rules.Input) (report.Report, error) {
	goRules, err := l.enabledGoRules()
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to get configured Go rules: %w", err)
	}

	goReport := report.Report{}

	for _, rule := range goRules {
		result, err := rule.Run(ctx, input)
		if err != nil {
			return report.Report{}, fmt.Errorf("error encountered in Go rule evaluation: %w", err)
		}

		goReport.Violations = append(goReport.Violations, result.Violations...)
	}

	return goReport, err
}

func (l Linter) prepareRegoArgs() []func(*rego.Rego) {
	var regoArgs []func(*rego.Rego)

	regoArgs = append(regoArgs,
		rego.ParsedQuery(query),
		rego.EnablePrintStatements(true),
		rego.PrintHook(topdown.NewPrintHook(os.Stderr)),
		rego.Function2(builtins.RegalParseModuleMeta, builtins.RegalParseModule),
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

	return regoArgs
}

func (l Linter) lintWithRegoRules(ctx context.Context, input rules.Input) (report.Report, error) {
	regoArgs := l.prepareRegoArgs()

	linterQuery, err := rego.New(regoArgs...).PrepareForEval(ctx)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed preparing query for linting: %w", err)
	}

	aggregate := report.Report{}

	for _, name := range input.FileNames {
		enhancedAST, err := parse.EnhanceAST(name, input.FileContent[name], input.Modules[name])
		if err != nil {
			return report.Report{}, fmt.Errorf("failed preparing AST: %w", err)
		}

		resultSet, err := linterQuery.Eval(ctx, rego.EvalInput(enhancedAST))
		if err != nil {
			return report.Report{}, fmt.Errorf("error encountered in query evaluation %w", err)
		}

		r, err := resultSetToReport(resultSet, input.FileContent[name])
		if err != nil {
			return report.Report{}, fmt.Errorf("failed to convert result set to report: %w", err)
		}

		aggregate.Violations = append(aggregate.Violations, r.Violations...)
	}

	return aggregate, nil
}

func resultSetToReport(resultSet rego.ResultSet, content string) (report.Report, error) {
	if len(resultSet) != 1 {
		return report.Report{}, fmt.Errorf("expected 1 item in resultset, got %d", len(resultSet))
	}

	r := report.Report{}
	if err := rio.JSONRoundTrip(resultSet[0].Bindings, &r); err != nil {
		return report.Report{},
			fmt.Errorf("JSON rountrip failed for bindings: %v %w", resultSet[0].Bindings, err)
	}

	for i, v := range r.Violations {
		r.Violations[i] = addText(v, content)
	}

	return r, nil
}

func addText(violation report.Violation, content string) report.Violation {
	if violation.Location.Text == nil && violation.Location.Row != 0 && violation.Location.Column != 0 {
		rowText, _ := rio.ReadRow(strings.NewReader(content), violation.Location.Row)
		violation.Location.Text = rowText
	}

	return violation
}

func (l Linter) mergedConfig() (config.Config, error) {
	if l.combinedConfig != nil {
		return *l.combinedConfig, nil
	}

	regalBundle, err := l.getBundleByName("regal")
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to get regal bundle: %w", err)
	}

	path := []string{"regal", "config", "provided"}

	bundled, err := util.SearchMap(regalBundle.Data, path)
	if err != nil {
		return config.Config{}, fmt.Errorf("config path not found %s: %w", strings.Join(path, "."), err)
	}

	bundledConf, ok := bundled.(map[string]any)
	if !ok {
		return config.Config{}, errors.New("expected 'rules' of object type")
	}

	userConfig := map[string]any{}

	if l.configBundle != nil && l.configBundle.Data != nil {
		userConfig = l.configBundle.Data[regalUserConfig].(map[string]any) //nolint:forcetypeassert
	}

	err = mergo.Merge(&bundledConf, userConfig)
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to merge config: %w", err)
	}

	return config.FromMap(bundledConf) //nolint:wrapcheck
}

func (l Linter) enabledGoRules() ([]rules.Rule, error) {
	allGoRules := []rules.Rule{
		rules.NewOpaFmtRule(),
	}

	conf, err := l.mergedConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create merged config: %w", err)
	}

	var enabledGoRules []rules.Rule

	for _, rule := range allGoRules {
		ruleConf, ok := conf.Rules[rule.Category()][rule.Name()]
		if ok && ruleConf.Enabled {
			enabledGoRules = append(enabledGoRules, rule)
		}
	}

	return enabledGoRules, nil
}

func (l Linter) getBundleByName(name string) (*bundle.Bundle, error) {
	if l.ruleBundles == nil {
		return nil, fmt.Errorf("no bundles loaded")
	}

	for _, ruleBundle := range l.ruleBundles {
		if metadataName, ok := ruleBundle.Manifest.Metadata["name"].(string); ok {
			if metadataName == name {
				return ruleBundle, nil
			}
		}
	}

	return nil, fmt.Errorf("no regal bundle found")
}
