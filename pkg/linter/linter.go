package linter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/gobwas/glob"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/metrics"
	"github.com/open-policy-agent/opa/profiler"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/open-policy-agent/opa/topdown/print"

	rbundle "github.com/styrainc/regal/bundle"
	rio "github.com/styrainc/regal/internal/io"
	regalmetrics "github.com/styrainc/regal/internal/metrics"
	"github.com/styrainc/regal/internal/parse"
	"github.com/styrainc/regal/internal/util"
	"github.com/styrainc/regal/pkg/builtins"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/report"
	"github.com/styrainc/regal/pkg/rules"
)

// Linter stores data to use for linting.
type Linter struct {
	customRuleFS         fs.FS
	printHook            print.Hook
	metrics              metrics.Metrics
	inputModules         *rules.Input
	userConfig           *config.Config
	combinedCfg          *config.Config
	dataBundle           *bundle.Bundle
	rootDir              string
	customRuleFSRootPath string
	inputPaths           []string
	ruleBundles          []*bundle.Bundle
	customRulesPaths     []string
	disable              []string
	disableCategory      []string
	enable               []string
	enableCategory       []string
	ignoreFiles          []string
	overriddenAggregates map[string][]report.Aggregate
	useCollectQuery      bool
	debugMode            bool
	exportAggregates     bool
	disableAll           bool
	enableAll            bool
	profiling            bool
}

//nolint:gochecknoglobals
var (
	// Single file provided as input.
	lintQuery = ast.MustParseBody(`lint := {
		"violations": data.regal.main.lint.violations,
		"notices": data.regal.main.lint.notices,
	}`)
	// More than one file provided as input.
	lintAndCollectQuery     = ast.MustParseBody("lint := data.regal.main.lint")
	lintWithAggregatesQuery = ast.MustParseBody("lint_aggregate := data.regal.main.lint_aggregate")
)

// NewLinter creates a new Regal linter.
func NewLinter() Linter {
	return Linter{
		ruleBundles: []*bundle.Bundle{&rbundle.LoadedBundle},
	}
}

// NewEmptyLinter creates a linter with no rule bundles.
func NewEmptyLinter() Linter {
	return Linter{}
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
func (l Linter) WithAddedBundle(b *bundle.Bundle) Linter {
	l.ruleBundles = append(l.ruleBundles, b)

	return l
}

// WithCustomRules adds custom rules for evaluation, from the Rego (and data) files provided at paths.
func (l Linter) WithCustomRules(paths []string) Linter {
	l.customRulesPaths = paths

	return l
}

// WithCustomRulesFromFS adds custom rules for evaluation from a filesystem implementing the fs.FS interface.
// A root path within the filesystem must also be specified. Note, _test.rego files will be ignored.
func (l Linter) WithCustomRulesFromFS(f fs.FS, rootPath string) Linter {
	l.customRuleFS = f
	l.customRuleFSRootPath = rootPath

	return l
}

// WithDebugMode enables debug mode.
func (l Linter) WithDebugMode(debugMode bool) Linter {
	l.debugMode = debugMode

	return l
}

// WithUserConfig provides config overrides set by the user.
func (l Linter) WithUserConfig(cfg config.Config) Linter {
	l.userConfig = &cfg

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

// WithMetrics enables metrics collection.
func (l Linter) WithMetrics(m metrics.Metrics) Linter {
	l.metrics = m

	return l
}

func (l Linter) WithPrintHook(printHook print.Hook) Linter {
	l.printHook = printHook

	return l
}

// WithProfiling enables profiling metrics.
func (l Linter) WithProfiling(enabled bool) Linter {
	l.profiling = enabled

	return l
}

// WithRootDir sets the root directory for the linter.
// A door directory or prefix can be used to resolve relative paths
// referenced in the linter configuration with absolute file paths or URIs.
func (l Linter) WithRootDir(rootDir string) Linter {
	l.rootDir = rootDir

	return l
}

// WithExportAggregates enables the setting of intermediate aggregate data
// on the final report. This is useful when you want to collect and
// aggregate state from multiple different linting runs.
func (l Linter) WithExportAggregates(enabled bool) Linter {
	l.exportAggregates = enabled

	return l
}

// WithCollectQuery forcibly enables the collect query even when there is
// only one file to lint.
func (l Linter) WithCollectQuery(enabled bool) Linter {
	l.useCollectQuery = enabled

	return l
}

// WithAggregates supplies aggregate data to a linter instance.
// Likely generated in a previous run, and used to provide a global context to
// a subsequent run of a single file lint.
func (l Linter) WithAggregates(aggregates map[string][]report.Aggregate) Linter {
	l.overriddenAggregates = aggregates

	return l
}

// Lint runs the linter on provided policies.
func (l Linter) Lint(ctx context.Context) (report.Report, error) {
	l.startTimer(regalmetrics.RegalLint)

	finalReport := report.Report{}

	if len(l.inputPaths) == 0 && l.inputModules == nil && len(l.overriddenAggregates) == 0 {
		return report.Report{}, errors.New("nothing provided to lint")
	}

	conf, err := l.GetConfig()
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to merge config: %w", err)
	}

	l.dataBundle = &bundle.Bundle{
		Manifest: bundle.Manifest{
			Roots:    &[]string{"internal"},
			Metadata: map[string]any{"name": "internal"},
		},
		Data: map[string]any{
			"internal": map[string]any{
				"combined_config": config.ToMap(*conf),
				"capabilities":    rio.ToMap(config.CapabilitiesForThisVersion()),
			},
		},
	}

	ignore := conf.Ignore.Files

	if len(l.ignoreFiles) > 0 {
		ignore = l.ignoreFiles
	}

	l.startTimer(regalmetrics.RegalFilterIgnoredFiles)

	filtered, err := config.FilterIgnoredPaths(l.inputPaths, ignore, true, l.rootDir)
	if err != nil {
		return report.Report{}, fmt.Errorf("errors encountered when reading files to lint: %w", err)
	}

	l.stopTimer(regalmetrics.RegalFilterIgnoredFiles)
	l.startTimer(regalmetrics.RegalInputParse)

	inputFromPaths, err := rules.InputFromPaths(filtered)
	if err != nil {
		return report.Report{}, fmt.Errorf("errors encountered when reading files to lint: %w", err)
	}

	l.stopTimer(regalmetrics.RegalInputParse)

	input := inputFromPaths

	if l.inputModules != nil {
		l.startTimer(regalmetrics.RegalFilterIgnoredModules)

		filteredPaths, err := config.FilterIgnoredPaths(
			l.inputModules.FileNames,
			ignore,
			false,
			l.rootDir,
		)
		if err != nil {
			return report.Report{}, fmt.Errorf("failed to filter paths: %w", err)
		}

		for _, filename := range filteredPaths {
			input.FileNames = append(input.FileNames, filename)
			input.Modules[filename] = l.inputModules.Modules[filename]
			input.FileContent[filename] = l.inputModules.FileContent[filename]
		}

		l.stopTimer(regalmetrics.RegalFilterIgnoredModules)
	}

	goReport, err := l.lintWithGoRules(ctx, input)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to lint using Go rules: %w", err)
	}

	finalReport.Violations = append(finalReport.Violations, goReport.Violations...)

	regoReport, err := l.lintWithRegoRules(ctx, input)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to lint using Rego rules: %w", err)
	}

	finalReport.Violations = append(finalReport.Violations, regoReport.Violations...)

	rulesSkippedCounter := 0

	for _, notice := range regoReport.Notices {
		if !slices.Contains(finalReport.Notices, notice) {
			finalReport.Notices = append(finalReport.Notices, notice)

			if notice.Severity != "none" {
				rulesSkippedCounter++
			}
		}
	}

	allAggregates := make(map[string][]report.Aggregate)

	if len(l.overriddenAggregates) > 0 {
		for k, aggregates := range l.overriddenAggregates {
			allAggregates[k] = append(allAggregates[k], aggregates...)
		}
	} else if len(input.FileNames) > 1 {
		for k, aggregates := range goReport.Aggregates {
			allAggregates[k] = append(allAggregates[k], aggregates...)
		}

		for k, aggregates := range regoReport.Aggregates {
			allAggregates[k] = append(allAggregates[k], aggregates...)
		}
	}

	if len(allAggregates) > 0 {
		aggregateReport, err := l.lintWithRegoAggregateRules(ctx, allAggregates, regoReport.IgnoreDirectives)
		if err != nil {
			return report.Report{}, fmt.Errorf("failed to lint using Rego aggregate rules: %w", err)
		}

		finalReport.Violations = append(finalReport.Violations, aggregateReport.Violations...)
	}

	finalReport.Summary = report.Summary{
		FilesScanned:  len(input.FileNames),
		FilesFailed:   len(finalReport.ViolationsFileCount()),
		RulesSkipped:  rulesSkippedCounter,
		NumViolations: len(finalReport.Violations),
	}

	if l.exportAggregates {
		finalReport.Aggregates = make(map[string][]report.Aggregate)
		for k, aggregates := range goReport.Aggregates {
			finalReport.Aggregates[k] = append(finalReport.Aggregates[k], aggregates...)
		}

		for k, aggregates := range regoReport.Aggregates {
			finalReport.Aggregates[k] = append(finalReport.Aggregates[k], aggregates...)
		}
	}

	if l.metrics != nil {
		l.metrics.Timer(regalmetrics.RegalLint).Stop()

		finalReport.Metrics = l.metrics.All()
	}

	if l.profiling {
		finalReport.AggregateProfile = regoReport.AggregateProfile
		finalReport.AggregateProfileToSortedProfile(10)
		finalReport.AggregateProfile = nil
	}

	return finalReport, nil
}

// DetermineEnabledRules returns the list of rules that are enabled based on
// the supplied configuration. This makes use of the Rego and Go rule settings
// to produce a single list of the rules that are to be run on this linter
// instance.
func (l Linter) DetermineEnabledRules(ctx context.Context) ([]string, error) {
	enabledRules := make([]string, 0)

	goRules, err := l.enabledGoRules()
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled Go rules: %w", err)
	}

	for _, rule := range goRules {
		enabledRules = append(enabledRules, rule.Name())
	}

	conf, err := l.GetConfig()
	if err != nil {
		return []string{}, fmt.Errorf("failed to merge config: %w", err)
	}

	l.dataBundle = &bundle.Bundle{
		Manifest: bundle.Manifest{
			Roots:    &[]string{"internal"},
			Metadata: map[string]any{"name": "internal"},
		},
		Data: map[string]any{
			"internal": map[string]any{
				"combined_config": config.ToMap(*conf),
				"capabilities":    rio.ToMap(config.CapabilitiesForThisVersion()),
			},
		},
	}

	queryStr := `[rule |
        data.regal.rules[cat][rule]
        data.regal.config.for_rule(cat, rule).level != "ignore"
    ]`

	query := ast.MustParseBody(queryStr)

	regoArgs, err := l.prepareRegoArgs(query)
	if err != nil {
		return nil, fmt.Errorf("failed preparing query %s: %w", queryStr, err)
	}

	rs, err := rego.New(regoArgs...).Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed evaluating query %s: %w", queryStr, err)
	}

	if len(rs) != 1 || len(rs[0].Expressions) != 1 {
		return nil, fmt.Errorf("expected exactly one expression, got %d", len(rs[0].Expressions))
	}

	list, ok := rs[0].Expressions[0].Value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected list, got %T", rs[0].Expressions[0].Value)
	}

	for _, item := range list {
		rule, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", item)
		}

		enabledRules = append(enabledRules, rule)
	}

	slices.Sort(enabledRules)

	return enabledRules, nil
}

// DetermineEnabledAggregateRules returns the list of aggregate rules that are
// enabled based on the configuration. This does not include any go rules.
func (l Linter) DetermineEnabledAggregateRules(ctx context.Context) ([]string, error) {
	enabledRules := make([]string, 0)

	conf, err := l.GetConfig()
	if err != nil {
		return []string{}, fmt.Errorf("failed to merge config: %w", err)
	}

	l.dataBundle = &bundle.Bundle{
		Manifest: bundle.Manifest{
			Roots:    &[]string{"internal"},
			Metadata: map[string]any{"name": "internal"},
		},
		Data: map[string]any{
			"internal": map[string]any{
				"combined_config": config.ToMap(*conf),
				"capabilities":    rio.ToMap(config.CapabilitiesForThisVersion()),
			},
		},
	}

	queryStr := `[rule |
        data.regal.rules[cat][rule].aggregate
        data.regal.config.for_rule(cat, rule).level != "ignore"
    ]`

	query := ast.MustParseBody(queryStr)

	regoArgs, err := l.prepareRegoArgs(query)
	if err != nil {
		return nil, fmt.Errorf("failed preparing query %s: %w", queryStr, err)
	}

	rs, err := rego.New(regoArgs...).Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed evaluating query %s: %w", queryStr, err)
	}

	if len(rs) != 1 || len(rs[0].Expressions) != 1 {
		return nil, fmt.Errorf("expected exactly one expression, got %d", len(rs[0].Expressions))
	}

	list, ok := rs[0].Expressions[0].Value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected list, got %T", rs[0].Expressions[0].Value)
	}

	for _, item := range list {
		rule, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", item)
		}

		enabledRules = append(enabledRules, rule)
	}

	slices.Sort(enabledRules)

	return enabledRules, nil
}

func (l Linter) lintWithGoRules(ctx context.Context, input rules.Input) (report.Report, error) {
	l.startTimer(regalmetrics.RegalLintGo)
	defer l.stopTimer(regalmetrics.RegalLintGo)

	goRules, err := l.enabledGoRules()
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to get configured Go rules: %w", err)
	}

	goReport := report.Report{}

	for _, rule := range goRules {
		inp, err := inputForRule(input, rule)
		if err != nil {
			return report.Report{}, fmt.Errorf("error encountered while filtering input files: %w", err)
		}

		result, err := rule.Run(ctx, inp)
		if err != nil {
			return report.Report{}, fmt.Errorf("error encountered in Go rule evaluation: %w", err)
		}

		goReport.Violations = append(goReport.Violations, result.Violations...)
	}

	return goReport, err
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

	return map[string]any{
		"eval": map[string]any{
			"params": params,
		},
	}
}

func (l Linter) prepareRegoArgs(query ast.Body) ([]func(*rego.Rego), error) {
	var regoArgs []func(*rego.Rego)

	roots := []string{"eval"}

	dataBundle := bundle.Bundle{
		Data:     l.paramsToRulesConfig(),
		Manifest: bundle.Manifest{Roots: &roots},
	}

	regoArgs = append(regoArgs,
		rego.Metrics(l.metrics),
		rego.ParsedQuery(query),
		rego.ParsedBundle("regal_eval_params", &dataBundle),
		rego.Function2(builtins.RegalParseModuleMeta, builtins.RegalParseModule),
		rego.Function1(builtins.RegalLastMeta, builtins.RegalLast),
	)

	if l.debugMode && l.printHook == nil {
		l.printHook = topdown.NewPrintHook(os.Stderr)
	}

	if l.printHook != nil {
		regoArgs = append(regoArgs,
			rego.EnablePrintStatements(true),
			rego.PrintHook(l.printHook),
		)
	}

	if l.dataBundle != nil {
		regoArgs = append(regoArgs, rego.ParsedBundle("internal", l.dataBundle))
	}

	if l.customRulesPaths != nil {
		regoArgs = append(regoArgs, rego.Load(l.customRulesPaths, rio.ExcludeTestFilter()))
	}

	if l.customRuleFS != nil && l.customRuleFSRootPath != "" {
		files, err := loadModulesFromCustomRuleFS(l.customRuleFS, l.customRuleFSRootPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load custom rules from FS: %w", err)
		}

		for path, content := range files {
			regoArgs = append(regoArgs, rego.Module(path, content))
		}
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

	return regoArgs, nil
}

func loadModulesFromCustomRuleFS(customRuleFS fs.FS, rootPath string) (map[string]string, error) {
	files := make(map[string]string)
	filter := rio.ExcludeTestFilter()

	err := fs.WalkDir(customRuleFS, rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk custom rule FS: %w", err)
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get info for custom rule file: %w", err)
		}

		if filter("", info, 0) {
			return nil
		}

		f, err := customRuleFS.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open custom rule file: %w", err)
		}
		defer f.Close()

		bs, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("failed to read custom rule file: %w", err)
		}

		files[path] = string(bs)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk custom rule FS: %w", err)
	}

	return files, nil
}

func (l Linter) lintWithRegoRules(ctx context.Context, input rules.Input) (report.Report, error) {
	l.startTimer(regalmetrics.RegalLintRego)
	defer l.stopTimer(regalmetrics.RegalLintRego)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var query ast.Body
	if len(input.FileNames) > 1 || l.useCollectQuery {
		query = lintAndCollectQuery
	} else {
		query = lintQuery
	}

	regoArgs, err := l.prepareRegoArgs(query)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed preparing query for linting: %w", err)
	}

	pq, err := rego.New(regoArgs...).PrepareForEval(ctx)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed preparing query for linting: %w", err)
	}

	regoReport := report.Report{}
	regoReport.Aggregates = make(map[string][]report.Aggregate)
	regoReport.IgnoreDirectives = make(map[string]map[string][]string)

	var wg sync.WaitGroup

	var mu sync.Mutex

	errCh := make(chan error)
	doneCh := make(chan bool)

	for _, name := range input.FileNames {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			enhancedAST, err := parse.PrepareAST(name, input.FileContent[name], input.Modules[name])
			if err != nil {
				errCh <- fmt.Errorf("failed preparing AST: %w", err)

				return
			}

			evalArgs := []rego.EvalOption{
				rego.EvalInput(enhancedAST),
			}

			if l.metrics != nil {
				evalArgs = append(evalArgs, rego.EvalMetrics(l.metrics))
			}

			var prof *profiler.Profiler
			if l.profiling {
				prof = profiler.New()
				evalArgs = append(evalArgs, rego.EvalQueryTracer(prof))
			}

			resultSet, err := pq.Eval(ctx, evalArgs...)
			if err != nil {
				errCh <- fmt.Errorf("error encountered in query evaluation %w", err)

				return
			}

			result, err := resultSetToReport(resultSet)
			if err != nil {
				errCh <- fmt.Errorf("failed to convert result set to report: %w", err)

				return
			}

			if l.profiling {
				// Perhaps we'll want to make this number configurable later, but do note that
				// this is only the top 10 locations for a *single* file, not the final report.
				profRep := prof.ReportTopNResults(10, []string{"total_time_ns"})

				result.AggregateProfile = make(map[string]report.ProfileEntry)

				for _, rs := range profRep {
					result.AggregateProfile[rs.Location.String()] = regalmetrics.FromExprStats(rs)
				}
			}

			mu.Lock()
			defer mu.Unlock()

			regoReport.Violations = append(regoReport.Violations, result.Violations...)
			regoReport.Notices = append(regoReport.Notices, result.Notices...)

			for k := range result.Aggregates {
				regoReport.Aggregates[k] = append(regoReport.Aggregates[k], result.Aggregates[k]...)
			}

			for k := range result.IgnoreDirectives {
				regoReport.IgnoreDirectives[k] = result.IgnoreDirectives[k]
			}

			if l.profiling {
				regoReport.AddProfileEntries(result.AggregateProfile)
			}
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
		return regoReport, nil
	}
}

func (l Linter) lintWithRegoAggregateRules(
	ctx context.Context,
	aggregates map[string][]report.Aggregate,
	ignoreDirectives map[string]map[string][]string,
) (report.Report, error) {
	l.startTimer(regalmetrics.RegalLintRegoAggregate)
	defer l.stopTimer(regalmetrics.RegalLintRegoAggregate)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	regoArgs, err := l.prepareRegoArgs(lintWithAggregatesQuery)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed preparing query for linting: %w", err)
	}

	pq, err := rego.New(regoArgs...).PrepareForEval(ctx)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed preparing query for linting: %w", err)
	}

	input := map[string]any{
		// This will be replaced by the routing policy to provide each
		// aggregate rule only the aggregated data from the same rule
		"aggregates_internal": aggregates,
		// There is no file provided in input here, but we'll provide *something* for
		// consistency, and to avoid silently failing with undefined should someone
		// refer to input.regal in an aggregate_report rule
		"ignore_directives": ignoreDirectives,
		"regal": map[string]any{
			"file": map[string]any{
				"name":  "__aggregate_report__",
				"lines": []string{},
			},
		},
	}

	evalArgs := []rego.EvalOption{rego.EvalInput(input)}

	if l.metrics != nil {
		evalArgs = append(evalArgs, rego.EvalMetrics(l.metrics))
	}

	resultSet, err := pq.Eval(ctx, evalArgs...)
	if err != nil {
		return report.Report{}, fmt.Errorf("error encountered in query evaluation %w", err)
	}

	result, err := resultSetToReport(resultSet)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to convert result set to report: %w", err)
	}

	for i := range result.Violations {
		result.Violations[i].IsAggregate = true
	}

	return result, nil
}

func resultSetToReport(resultSet rego.ResultSet) (report.Report, error) {
	if len(resultSet) != 1 {
		return report.Report{}, fmt.Errorf("expected 1 item in resultset, got %d", len(resultSet))
	}

	r := report.Report{}

	if binding, ok := resultSet[0].Bindings["lint"]; ok {
		if err := rio.JSONRoundTrip(binding, &r); err != nil {
			return report.Report{}, fmt.Errorf("JSON rountrip failed for bindings: %v %w", binding, err)
		}
	}

	if binding, ok := resultSet[0].Bindings["lint_aggregate"]; ok {
		if err := rio.JSONRoundTrip(binding, &r); err != nil {
			return report.Report{}, fmt.Errorf("JSON rountrip failed for bindings: %v %w", binding, err)
		}
	}

	return r, nil
}

// GetConfig returns the final configuration for the linter, i.e. Regal's default
// configuration plus any user-provided configuration merged on top of it.
func (l Linter) GetConfig() (*config.Config, error) {
	if l.combinedCfg != nil {
		return l.combinedCfg, nil
	}

	regalBundle, err := l.getBundleByName("regal")
	if err != nil {
		return &config.Config{}, fmt.Errorf("failed to get regal bundle: %w", err)
	}

	mergedConf, err := config.LoadConfigWithDefaultsFromBundle(regalBundle, l.userConfig)
	if err != nil {
		return &config.Config{}, fmt.Errorf("failed to read provided config: %w", err)
	}

	if l.debugMode {
		bs, err := yaml.Marshal(mergedConf)
		if err != nil {
			return &config.Config{}, fmt.Errorf("failed to marshal config: %w", err)
		}

		log.Println("merged provided and user config:")
		log.Println(string(bs))
	}

	l.combinedCfg = &mergedConf

	return l.combinedCfg, nil
}

func (l Linter) enabledGoRules() ([]rules.Rule, error) {
	var enabledGoRules []rules.Rule

	// enabling/disabling all rules takes precedence and entirely disregards configuration
	// files, but still respects the enable/disable category or rule flags

	if l.disableAll {
		for _, rule := range rules.AllGoRules(config.Config{}) {
			if slices.Contains(l.enableCategory, rule.Category()) || slices.Contains(l.enable, rule.Name()) {
				enabledGoRules = append(enabledGoRules, rule)
			}
		}

		return enabledGoRules, nil
	}

	if l.enableAll {
		for _, rule := range rules.AllGoRules(config.Config{}) {
			if !slices.Contains(l.disableCategory, rule.Category()) && !slices.Contains(l.disable, rule.Name()) {
				enabledGoRules = append(enabledGoRules, rule)
			}
		}

		return enabledGoRules, nil
	}

	conf, err := l.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create merged config: %w", err)
	}

	for _, rule := range rules.AllGoRules(*conf) {
		// disabling specific rule has the highest precedence
		if slices.Contains(l.disable, rule.Name()) {
			continue
		}

		// likewise for enabling specific rule
		if slices.Contains(l.enable, rule.Name()) {
			enabledGoRules = append(enabledGoRules, rule)

			continue
		}

		// next highest precedence is disabling / enabling a category
		if slices.Contains(l.disableCategory, rule.Category()) {
			continue
		}

		if slices.Contains(l.enableCategory, rule.Category()) {
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
		return nil, errors.New("no bundles loaded")
	}

	for _, ruleBundle := range l.ruleBundles {
		if metadataName, ok := ruleBundle.Manifest.Metadata["name"].(string); ok {
			if metadataName == name {
				return ruleBundle, nil
			}
		}
	}

	return nil, errors.New("no regal bundle found")
}

func (l Linter) startTimer(name string) {
	if l.metrics != nil {
		l.metrics.Timer(name).Start()
	}
}

func (l Linter) stopTimer(name string) {
	if l.metrics != nil {
		l.metrics.Timer(name).Stop()
	}
}
