package linter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"dario.cat/mergo"
	"github.com/gobwas/glob"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"

	rbundle "github.com/styrainc/regal/bundle"
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
	inputPaths       []string
	inputModules     *rules.Input
	ruleBundles      []*bundle.Bundle
	configBundle     *bundle.Bundle
	customRulesPaths []string
	combinedConfig   *config.Config
	debugMode        bool
	disable          []string
	disableAll       bool
	disableCategory  []string
	enable           []string
	enableAll        bool
	enableCategory   []string
	ignoreFiles      []string
}

const regalUserConfig = "regal_user_config"

// NewLinter creates a new Regal linter.
func NewLinter() Linter {
	regalRules := rio.MustLoadRegalBundleFS(rbundle.Bundle)

	return Linter{
		ruleBundles: []*bundle.Bundle{&regalRules},
	}
}

// WithInputPaths sets the inputPaths to lint. Note that these will be
// filtered according to the ignore options.
func (l Linter) WithInputPaths(paths []string) Linter {
	l.inputPaths = paths

	return l
}

// WithInputModules sets the input modules to lint. This is used for programmatic
// access, where you don't necessarily want to lint *files*.
func (l Linter) WithInputModules(input *rules.Input) Linter {
	l.inputModules = input

	return l
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

// WithDebugMode enables debug mode.
func (l Linter) WithDebugMode(debugMode bool) Linter {
	l.debugMode = debugMode

	return l
}

// WithUserConfig provides config overrides set by the user.
func (l Linter) WithUserConfig(cfg config.Config) Linter {
	l.configBundle = &bundle.Bundle{
		Manifest: bundle.Manifest{
			Roots:    &[]string{regalUserConfig},
			Metadata: map[string]any{"name": regalUserConfig},
		},
		Data: map[string]any{regalUserConfig: config.ToMap(cfg)},
	}

	return l
}

// WithDisabledRules disables provided rules. This overrides configuration provided in file.
func (l Linter) WithDisabledRules(disable ...string) Linter {
	l.disable = disable

	return l
}

// WithDisableAll disables all rules when set to true. This overrides configuration provided in file.
func (l Linter) WithDisableAll(disableAll bool) Linter {
	l.disableAll = disableAll

	return l
}

// WithDisabledCategories disables provided categories of rules. This overrides configuration provided in file.
func (l Linter) WithDisabledCategories(disableCategory ...string) Linter {
	l.disableCategory = disableCategory

	return l
}

// WithEnabledRules enables provided rules. This overrides configuration provided in file.
func (l Linter) WithEnabledRules(enable ...string) Linter {
	l.enable = enable

	return l
}

// WithEnableAll enables all rules when set to true. This overrides configuration provided in file.
func (l Linter) WithEnableAll(enableAll bool) Linter {
	l.enableAll = enableAll

	return l
}

// WithEnabledCategories enables provided categories of rules. This overrides configuration provided in file.
func (l Linter) WithEnabledCategories(enableCategory ...string) Linter {
	l.enableCategory = enableCategory

	return l
}

// WithIgnore excludes files matching patterns. This overrides configuration provided in file.
func (l Linter) WithIgnore(ignore []string) Linter {
	l.ignoreFiles = ignore

	return l
}

var query = ast.MustParseBody("violations = data.regal.main.report") //nolint:gochecknoglobals

// Lint runs the linter on provided policies.
func (l Linter) Lint(ctx context.Context) (report.Report, error) {
	aggregate := report.Report{}

	if len(l.inputPaths) == 0 && l.inputModules == nil {
		return report.Report{}, errors.New("nothing provided to lint")
	}

	conf, err := l.mergedConfig()
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to merge config: %w", err)
	}

	l.combinedConfig = &conf

	ignore := conf.Ignore.Files

	if len(l.ignoreFiles) > 0 {
		ignore = l.ignoreFiles
	}

	filtered, err := config.FilterIgnoredPaths(l.inputPaths, ignore, true)
	if err != nil {
		return report.Report{}, fmt.Errorf("errors encountered when reading files to lint: %w", err)
	}

	inputFromPaths, err := rules.InputFromPaths(filtered)
	if err != nil {
		return report.Report{}, fmt.Errorf("errors encountered when reading files to lint: %w", err)
	}

	input := inputFromPaths

	if l.inputModules != nil {
		filteredPaths, err := config.FilterIgnoredPaths(l.inputModules.FileNames, ignore, false)
		if err != nil {
			return report.Report{}, fmt.Errorf("failed to filter paths: %w", err)
		}

		for _, filename := range filteredPaths {
			input.FileNames = append(input.FileNames, filename)
			input.Modules[filename] = l.inputModules.Modules[filename]
			input.FileContent[filename] = l.inputModules.FileContent[filename]
		}
	}

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

	aggregate.Summary = report.Summary{
		FilesScanned:  len(input.FileNames),
		FilesFailed:   len(aggregate.ViolationsFileCount()),
		FilesSkipped:  0,
		NumViolations: len(aggregate.Violations),
	}

	return aggregate, nil
}

func (l Linter) lintWithGoRules(ctx context.Context, input rules.Input) (report.Report, error) {
	goRules, err := l.enabledGoRules()
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to get configured Go rules: %w", err)
	}

	aggregate := report.Report{}

	for _, rule := range goRules {
		inp, err := inputForRule(input, rule)
		if err != nil {
			return report.Report{}, fmt.Errorf("error encountered while filtering input files: %w", err)
		}

		result, err := rule.Run(ctx, inp)
		if err != nil {
			return report.Report{}, fmt.Errorf("error encountered in Go rule evaluation: %w", err)
		}

		aggregate.Violations = append(aggregate.Violations, result.Violations...)
	}

	return aggregate, err
}

func inputForRule(input rules.Input, rule rules.Rule) (rules.Input, error) {
	ignore := rule.Config().Ignore

	var ignoreFiles []string

	if ignore != nil {
		ignoreFiles = ignore.Files
	}

	return filterInputFiles(input, ignoreFiles)
}

func filterInputFiles(input rules.Input, ignore []string) (rules.Input, error) {
	if len(ignore) == 0 {
		return input, nil
	}

	n := len(input.FileNames)
	newInput := rules.Input{
		FileNames:   make([]string, 0, n),
		FileContent: make(map[string]string, n),
		Modules:     make(map[string]*ast.Module, n),
	}

outer:
	for _, f := range input.FileNames {
		for _, pattern := range ignore {
			if pattern == "" {
				continue
			}

			excluded, err := excludeFile(pattern, f)
			if err != nil {
				return rules.Input{}, fmt.Errorf("failed to check for exclusion using pattern %s: %w", pattern, err)
			}

			if excluded {
				continue outer
			}
		}

		newInput.FileNames = append(newInput.FileNames, f)
		newInput.FileContent[f] = input.FileContent[f]
		newInput.Modules[f] = input.Modules[f]
	}

	return newInput, nil
}

// excludeFile imitates the pattern matching of .gitignore files
// See `exclusion.rego` for details on the implementation.
func excludeFile(pattern string, filename string) (bool, error) {
	n := len(pattern)

	// Internal slashes means path is relative to root, otherwise it can
	// appear anywhere in the directory (--> **/)
	if !strings.Contains(pattern[:n-1], "/") {
		pattern = "**/" + pattern
	}

	// Leading slash?
	pattern = strings.TrimPrefix(pattern, "/")

	// Leading double-star?
	var ps []string
	if strings.HasPrefix(pattern, "**/") {
		ps = []string{pattern, strings.TrimPrefix(pattern, "**/")}
	} else {
		ps = []string{pattern}
	}

	var ps1 []string

	// trailing slash?
	for _, p := range ps {
		switch {
		case strings.HasSuffix(p, "/"):
			ps1 = append(ps1, p+"**")
		case !strings.HasSuffix(p, "/") && !strings.HasSuffix(p, "**"):
			ps1 = append(ps1, p, p+"/**")
		default:
			ps1 = append(ps1, p)
		}
	}

	// Loop through patterns and return true on first match
	for _, p := range ps1 {
		g, err := glob.Compile(p, '/')
		if err != nil {
			return false, fmt.Errorf("failed to compile pattern %s: %w", p, err)
		}

		if g.Match(filename) {
			return true, nil
		}
	}

	return false, nil
}

func (l Linter) paramsToRulesConfig() map[string]any {
	params := map[string]any{
		"disable_all":      l.disableAll,
		"disable_category": util.NullToEmpty(l.disableCategory),
		"disable":          util.NullToEmpty(l.disable),
		"enable_all":       l.enableAll,
		"enable_category":  util.NullToEmpty(l.enableCategory),
		"enable":           util.NullToEmpty(l.enable),
	}

	if l.ignoreFiles != nil {
		params["ignore_files"] = l.ignoreFiles
	}

	return map[string]interface{}{
		"eval": map[string]any{
			"params": params,
		},
	}
}

func (l Linter) prepareRegoArgs() []func(*rego.Rego) {
	var regoArgs []func(*rego.Rego)

	roots := []string{"eval"}

	dataBundle := bundle.Bundle{
		Data:     l.paramsToRulesConfig(),
		Manifest: bundle.Manifest{Roots: &roots},
	}

	regoArgs = append(regoArgs,
		rego.ParsedQuery(query),
		rego.ParsedBundle("regal_eval_params", &dataBundle),
		rego.Function2(builtins.RegalParseModuleMeta, builtins.RegalParseModule),
		rego.Function1(builtins.RegalJSONPrettyMeta, builtins.RegalJSONPretty),
		rego.Function1(builtins.RegalLastMeta, builtins.RegalLast),
	)

	if l.debugMode {
		regoArgs = append(regoArgs,
			rego.EnablePrintStatements(true),
			rego.PrintHook(topdown.NewPrintHook(os.Stderr)),
		)
	}

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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	regoArgs := l.prepareRegoArgs()

	linterQuery, err := rego.New(regoArgs...).PrepareForEval(ctx)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed preparing query for linting: %w", err)
	}

	aggregate := report.Report{}

	var wg sync.WaitGroup

	var mu sync.Mutex

	errCh := make(chan error)
	doneCh := make(chan bool)

	for _, name := range input.FileNames {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			enhancedAST, err := parse.EnhanceAST(name, input.FileContent[name], input.Modules[name])
			if err != nil {
				errCh <- fmt.Errorf("failed preparing AST: %w", err)

				return
			}

			resultSet, err := linterQuery.Eval(ctx, rego.EvalInput(enhancedAST))
			if err != nil {
				errCh <- fmt.Errorf("error encountered in query evaluation %w", err)

				return
			}

			result, err := resultSetToReport(resultSet)
			if err != nil {
				errCh <- fmt.Errorf("failed to convert result set to report: %w", err)

				return
			}

			mu.Lock()
			aggregate.Violations = append(aggregate.Violations, result.Violations...)
			mu.Unlock()
		}(name)
	}

	go func() {
		wg.Wait()
		doneCh <- true
	}()

	select {
	case <-ctx.Done():
		return report.Report{}, fmt.Errorf("context cancelled: %w", ctx.Err())
	case err := <-errCh:
		return report.Report{}, fmt.Errorf("error encountered in rule evaluation %w", err)
	case <-doneCh:
		return aggregate, nil
	}
}

func resultSetToReport(resultSet rego.ResultSet) (report.Report, error) {
	if len(resultSet) != 1 {
		return report.Report{}, fmt.Errorf("expected 1 item in resultset, got %d", len(resultSet))
	}

	r := report.Report{}
	if err := rio.JSONRoundTrip(resultSet[0].Bindings, &r); err != nil {
		return report.Report{},
			fmt.Errorf("JSON rountrip failed for bindings: %v %w", resultSet[0].Bindings, err)
	}

	return r, nil
}

func (l Linter) readProvidedConfig() (config.Config, error) {
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

	return config.FromMap(bundledConf) //nolint:wrapcheck
}

func (l Linter) readUserConfig() (config.Config, error) {
	if l.configBundle != nil && l.configBundle.Data != nil {
		userConfigMap := l.configBundle.Data[regalUserConfig].(map[string]any) //nolint:forcetypeassert

		userConfig, err := config.FromMap(userConfigMap)
		if err != nil {
			return config.Config{}, fmt.Errorf("failed to create config from map: %w", err)
		}

		return userConfig, nil
	}

	return config.Config{}, nil
}

func (l Linter) mergedConfig() (config.Config, error) {
	if l.combinedConfig != nil {
		return *l.combinedConfig, nil
	}

	mergedConf, err := l.readProvidedConfig()
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to read provided config: %w", err)
	}

	ruleLevels := providedConfLevels(mergedConf)

	userConf, err := l.readUserConfig()
	if err != nil {
		// Note that user config is optional — this will only happen if config is somehow invalid
		return config.Config{}, fmt.Errorf("failed to read user config: %w", err)
	}

	err = mergo.Merge(&mergedConf, userConf, mergo.WithOverride)
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to merge config: %w", err)
	}

	// If the user configuration contains rules with the level unset, copy the level from the provided configuration
	for categoryName, rulesByCategory := range mergedConf.Rules {
		for ruleName, rule := range rulesByCategory {
			if rule.Level == "" {
				if providedLevel, ok := ruleLevels[ruleName]; ok {
					rule.Level = providedLevel
					mergedConf.Rules[categoryName][ruleName] = rule
				}
			}
		}
	}

	if l.debugMode {
		bs, err := yaml.Marshal(mergedConf)
		if err != nil {
			return config.Config{}, fmt.Errorf("failed to marshal config: %w", err)
		}

		log.Println("merged provided and user config:")
		log.Println(string(bs))
	}

	return mergedConf, nil
}

// Copy the level of each rule from the provided configuration.
func providedConfLevels(conf config.Config) map[string]string {
	ruleLevels := make(map[string]string)

	for categoryName, rulesByCategory := range conf.Rules {
		for ruleName := range rulesByCategory {
			// Note that this assumes all rules have unique names,
			// which we'll likely always want for provided rules
			ruleLevels[ruleName] = conf.Rules[categoryName][ruleName].Level
		}
	}

	return ruleLevels
}

func (l Linter) enabledGoRules() ([]rules.Rule, error) {
	var enabledGoRules []rules.Rule

	// enabling/disabling all rules takes precedence and entirely disregards configuration
	// files, but still respects the enable/disable category or rule flags

	if l.disableAll {
		for _, rule := range rules.AllGoRules(config.Config{}) {
			if util.Contains(l.enableCategory, rule.Category()) || util.Contains(l.enable, rule.Name()) {
				enabledGoRules = append(enabledGoRules, rule)
			}
		}

		return enabledGoRules, nil
	}

	if l.enableAll {
		for _, rule := range rules.AllGoRules(config.Config{}) {
			if !util.Contains(l.disableCategory, rule.Category()) && !util.Contains(l.disable, rule.Name()) {
				enabledGoRules = append(enabledGoRules, rule)
			}
		}

		return enabledGoRules, nil
	}

	conf, err := l.mergedConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create merged config: %w", err)
	}

	for _, rule := range rules.AllGoRules(conf) {
		// disabling specific rule has the highest precedence
		if util.Contains(l.disable, rule.Name()) {
			continue
		}

		// likewise for enabling specific rule
		if util.Contains(l.enable, rule.Name()) {
			enabledGoRules = append(enabledGoRules, rule)

			continue
		}

		// next highest precedence is disabling / enabling a category
		if util.Contains(l.disableCategory, rule.Category()) {
			continue
		}

		if util.Contains(l.enableCategory, rule.Category()) {
			enabledGoRules = append(enabledGoRules, rule)

			continue
		}

		// if none of the above applies, check the config for the rule
		if rule.Config().Level != "ignore" {
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
