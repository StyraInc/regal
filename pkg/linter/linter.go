package linter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing/fstest"

	"github.com/anderseknert/roast/pkg/encoding"
	"github.com/anderseknert/roast/pkg/transform"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/bundle"
	"github.com/open-policy-agent/opa/v1/metrics"
	"github.com/open-policy-agent/opa/v1/profiler"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/topdown"
	"github.com/open-policy-agent/opa/v1/topdown/print"
	outil "github.com/open-policy-agent/opa/v1/util"

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
	printHook            print.Hook
	metrics              metrics.Metrics
	inputModules         *rules.Input
	userConfig           *config.Config
	combinedCfg          *config.Config
	dataBundle           *bundle.Bundle
	pathPrefix           string
	customRuleError      error
	inputPaths           []string
	ruleBundles          []*bundle.Bundle
	disable              []string
	disableCategory      []string
	enable               []string
	enableCategory       []string
	ignoreFiles          []string
	customRuleModules    []*ast.Module
	overriddenAggregates map[string][]report.Aggregate
	useCollectQuery      bool
	debugMode            bool
	exportAggregates     bool
	disableAll           bool
	enableAll            bool
	profiling            bool
	instrumentation      bool
	hasCustomRules       bool
}

//nolint:gochecknoglobals
var lintQuery = ast.MustParseBody("lint := data.regal.main.lint")

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
	for _, path := range paths {
		stat, err := os.Stat(path)
		if err != nil {
			l.customRuleError = fmt.Errorf("failed to stat custom rule file %s: %w", path, err)

			return l
		}

		if stat.IsDir() {
			l = l.WithCustomRulesFromFS(os.DirFS(path), ".")
		} else {
			contents, err := os.ReadFile(path)
			if err != nil {
				l.customRuleError = fmt.Errorf("failed to read custom rule file %s: %w", path, err)

				return l
			}

			l = l.WithCustomRulesFromFS(fstest.MapFS{
				filepath.Base(path): &fstest.MapFile{Data: contents},
			}, ".")
		}
	}

	return l
}

// WithCustomRulesFromFS adds custom rules for evaluation from a filesystem implementing the fs.FS interface.
// A root path within the filesystem must also be specified. Note, _test.rego files will be ignored.
func (l Linter) WithCustomRulesFromFS(f fs.FS, rootPath string) Linter {
	if f == nil {
		return l
	}

	l.hasCustomRules = true

	modules, err := loadModulesFromCustomRuleFS(f, rootPath)
	if err != nil {
		l.customRuleError = err

		return l
	}

	for _, m := range modules {
		l.customRuleModules = append(l.customRuleModules, m)
	}

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

// WithInstrumentation enables instrumentation metrics.
func (l Linter) WithInstrumentation(enabled bool) Linter {
	l.instrumentation = enabled

	return l
}

// WithPathPrefix sets the root path prefix for the linter.
// A root directory prefix can be used to resolve relative paths
// referenced in the linter configuration with absolute file paths or URIs.
func (l Linter) WithPathPrefix(pathPrefix string) Linter {
	l.pathPrefix = pathPrefix

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

	err := l.validate()
	if err != nil {
		return report.Report{}, fmt.Errorf("validation failed: %w", err)
	}

	conf, err := l.GetConfig()
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to merge config: %w", err)
	}

	l.dataBundle = l.createDataBundle(*conf)

	ignore := conf.Ignore.Files

	if len(l.ignoreFiles) > 0 {
		ignore = l.ignoreFiles
	}

	l.startTimer(regalmetrics.RegalFilterIgnoredFiles)

	filtered, err := config.FilterIgnoredPaths(l.inputPaths, ignore, true, l.pathPrefix)
	if err != nil {
		return report.Report{}, fmt.Errorf("errors encountered when reading files to lint: %w", err)
	}

	l.stopTimer(regalmetrics.RegalFilterIgnoredFiles)
	l.startTimer(regalmetrics.RegalInputParse)

	var versionsMap map[string]ast.RegoVersion

	if l.pathPrefix != "" && !strings.HasPrefix(l.pathPrefix, "file://") {
		versionsMap, err = config.AllRegoVersions(l.pathPrefix, conf)
		if err != nil && l.debugMode {
			log.Printf("failed to get configured Rego versions: %v", err)
		}
	}

	inputFromPaths, err := rules.InputFromPaths(filtered, l.pathPrefix, versionsMap)
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
			l.pathPrefix,
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

	regoArgs, err := l.prepareRegoArgs(lintQuery)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed preparing query for linting: %w", err)
	}

	pq, err := rego.New(regoArgs...).PrepareForEval(ctx)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed preparing query for linting: %w", err)
	}

	regoReport, err := l.lintWithRegoRules(ctx, &pq, input)
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
		for k, aggregates := range regoReport.Aggregates {
			allAggregates[k] = append(allAggregates[k], aggregates...)
		}
	}

	if len(allAggregates) > 0 {
		aggregateReport, err := l.lintWithRegoAggregateRules(ctx, &pq, allAggregates, regoReport.IgnoreDirectives)
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

func (l Linter) validate() error {
	if len(l.inputPaths) == 0 && l.inputModules == nil && len(l.overriddenAggregates) == 0 {
		return errors.New("nothing provided to lint")
	}

	if l.customRuleError != nil {
		return fmt.Errorf("failed to load custom rules: %w", l.customRuleError)
	}

	conf, err := l.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to merge config: %w", err)
	}

	validCategories := util.NewSet[string]()
	validRules := util.NewSet[string]()

	// Add all built-in rules
	for _, b := range l.ruleBundles {
		for _, module := range b.Modules {
			parts := util.UnquotedPath(module.Parsed.Package.Path)
			// 1     2     3   4
			// regal.rules.cat.rule
			if len(parts) != 4 {
				continue
			}

			validCategories.Add(parts[2])
			validRules.Add(parts[3])
		}
	}

	// Add any custom rules
	for _, module := range l.customRuleModules {
		parts := util.UnquotedPath(module.Package.Path)
		// 1      2     3     4   5
		// custom.regal.rules.cat.rule
		if len(parts) != 5 {
			continue
		}

		validCategories.Add(parts[3])
		validRules.Add(parts[4])
	}

	configuredCategories := util.NewSet(outil.Keys(conf.Rules)...)
	configuredRules := util.NewSet[string]()

	for _, cat := range conf.Rules {
		configuredRules.Add(outil.Keys(cat)...)
	}

	configuredRules.Add(l.enable...)
	configuredRules.Add(l.disable...)
	configuredCategories.Add(l.enableCategory...)
	configuredCategories.Add(l.disableCategory...)

	var invalidCategories []string

	if diff := configuredCategories.Diff(validCategories); diff.Size() > 0 {
		invalidCategories = diff.Items()
	}

	var invalidRules []string

	if diff := configuredRules.Diff(validRules); diff.Size() > 0 {
		invalidRules = diff.Items()
	}

	switch {
	case len(invalidCategories) > 0 && len(invalidRules) > 0:
		return fmt.Errorf("unknown categories: %v, unknown rules: %v", invalidCategories, invalidRules)
	case len(invalidCategories) > 0:
		return fmt.Errorf("unknown categories: %v", invalidCategories)
	case len(invalidRules) > 0:
		return fmt.Errorf("unknown rules: %v", invalidRules)
	}

	return nil
}

// DetermineEnabledRules returns the list of rules that are enabled based on
// the supplied configuration. This makes use of the Rego and Go rule settings
// to produce a single list of the rules that are to be run on this linter
// instance.
func (l Linter) DetermineEnabledRules(ctx context.Context) ([]string, error) {
	conf, err := l.GetConfig()
	if err != nil {
		return []string{}, fmt.Errorf("failed to merge config: %w", err)
	}

	l.dataBundle = l.createDataBundle(*conf)

	queryStr := `[rule |
        data.regal.rules[cat][rule]
		object.get(data.regal.rules[cat][rule], "notices", set()) == set()
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

	enabledRules := make([]string, 0)

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

	l.dataBundle = l.createDataBundle(*conf)

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

func (l Linter) createDataBundle(conf config.Config) *bundle.Bundle {
	params := map[string]any{
		"disable_all":      l.disableAll,
		"disable_category": util.NullToEmpty(l.disableCategory),
		"disable":          util.NullToEmpty(l.disable),
		"enable_all":       l.enableAll,
		"enable_category":  util.NullToEmpty(l.enableCategory),
		"enable":           util.NullToEmpty(l.enable),
		"ignore_files":     util.NullToEmpty(l.ignoreFiles),
	}

	return &bundle.Bundle{
		Manifest: bundle.Manifest{
			Roots:    &[]string{"internal", "eval"},
			Metadata: map[string]any{"name": "internal"},
		},
		Data: map[string]any{
			"eval": map[string]any{
				"params": params,
			},
			"internal": map[string]any{
				"combined_config": config.ToMap(conf),
				"capabilities":    rio.ToMap(config.CapabilitiesForThisVersion()),
				"path_prefix":     l.pathPrefix,
			},
		},
	}
}

func (l Linter) prepareRegoArgs(query ast.Body) ([]func(*rego.Rego), error) {
	regoArgs := append([]func(*rego.Rego){
		rego.StoreReadAST(true),
		rego.Metrics(l.metrics),
		rego.ParsedQuery(query),
	}, builtins.RegalBuiltinRegoFuncs...)

	if l.debugMode && l.printHook == nil {
		l.printHook = topdown.NewPrintHook(os.Stderr)
	}

	if l.printHook != nil {
		regoArgs = append(regoArgs,
			rego.EnablePrintStatements(true),
			rego.PrintHook(l.printHook),
		)
	}

	if l.instrumentation {
		regoArgs = append(regoArgs, rego.Instrument(true))
	}

	if l.dataBundle != nil {
		regoArgs = append(regoArgs, rego.ParsedBundle("internal", l.dataBundle))
	}

	if l.hasCustomRules {
		if l.customRuleError != nil {
			return nil, fmt.Errorf("failed to load custom rules: %w", l.customRuleError)
		}

		for _, m := range l.customRuleModules {
			regoArgs = append(regoArgs, rego.ParsedModule(m))
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

func loadModulesFromCustomRuleFS(customRuleFS fs.FS, rootPath string) (map[string]*ast.Module, error) {
	modules := make(map[string]*ast.Module)
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

		m, err := ast.ParseModule(path, outil.ByteSliceToString(bs))
		if err != nil {
			return fmt.Errorf("failed to parse custom rule file %q: %w", path, err)
		}

		modules[path] = m

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk custom rule FS: %w", err)
	}

	return modules, nil
}

func (l Linter) lintWithRegoRules(
	ctx context.Context,
	pq *rego.PreparedEvalQuery,
	input rules.Input,
) (report.Report, error) {
	l.startTimer(regalmetrics.RegalLintRego)
	defer l.stopTimer(regalmetrics.RegalLintRego)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	regoReport := report.Report{}
	regoReport.Aggregates = make(map[string][]report.Aggregate)
	regoReport.IgnoreDirectives = make(map[string]map[string][]string)

	var operationCollect bool
	if len(input.FileNames) > 1 || l.useCollectQuery {
		operationCollect = true
	}

	var wg sync.WaitGroup

	var mu sync.Mutex

	// the error channel is buffered to prevent blocking
	// caused by the context cancellation happening before
	// errors are sent and the per-file goroutines can exit.
	errCh := make(chan error, len(input.FileNames))
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

			regalInput, ok := enhancedAST["regal"].(map[string]any)
			if ok {
				operations := []string{"lint"}

				if operationCollect {
					operations = append(operations, "collect")
				}

				regalInput["operations"] = operations
			}

			inputValue, err := transform.ToOPAInputValue(enhancedAST)
			if err != nil {
				errCh <- fmt.Errorf("failed to transform input value: %w", err)

				return
			}

			evalArgs := []rego.EvalOption{
				rego.EvalParsedInput(inputValue),
			}

			if l.metrics != nil {
				evalArgs = append(evalArgs, rego.EvalMetrics(l.metrics))
			}

			var prof *profiler.Profiler
			if l.profiling {
				prof = profiler.New()
				evalArgs = append(evalArgs, rego.EvalQueryTracer(prof))
			}

			if l.instrumentation {
				evalArgs = append(evalArgs, rego.EvalInstrument(true))
			}

			resultSet, err := pq.Eval(ctx, evalArgs...)
			if err != nil {
				errCh <- fmt.Errorf("error encountered in query evaluation %w", err)

				return
			}

			result, err := resultSetToReport(resultSet, false)
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
				// Custom aggregate rules that have been invoked but not returned any data
				// will return an empty map to signal that they have been called, and that
				// the aggregate report for this rule should be invoked even when no data
				// was aggregated. This because the absence of data is exactly what some rules
				// will want to report on.
				for _, agg := range result.Aggregates[k] {
					if len(agg) == 0 {
						if _, ok := regoReport.Aggregates[k]; !ok {
							regoReport.Aggregates[k] = make([]report.Aggregate, 0)
						}
					} else {
						regoReport.Aggregates[k] = append(regoReport.Aggregates[k], agg)
					}
				}
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
	pq *rego.PreparedEvalQuery,
	aggregates map[string][]report.Aggregate,
	ignoreDirectives map[string]map[string][]string,
) (report.Report, error) {
	l.startTimer(regalmetrics.RegalLintRegoAggregate)
	defer l.stopTimer(regalmetrics.RegalLintRegoAggregate)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	input := map[string]any{
		// This will be replaced by the routing policy to provide each
		// aggregate rule only the aggregated data from the same rule
		"aggregates_internal": aggregates,
		// There is no file provided in input here, but we'll provide *something* for
		// consistency, and to avoid silently failing with undefined should someone
		// refer to input.regal in an aggregate_report rule
		"ignore_directives": ignoreDirectives,
		"regal": map[string]any{
			"operations": []string{"aggregate"},
			"file": map[string]any{
				"name":  "__aggregate_report__",
				"lines": []string{},
			},
		},
	}

	inputValue, err := transform.ToOPAInputValue(input)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to transform input value: %w", err)
	}

	evalArgs := []rego.EvalOption{rego.EvalParsedInput(inputValue)}

	if l.metrics != nil {
		evalArgs = append(evalArgs, rego.EvalMetrics(l.metrics))
	}

	resultSet, err := pq.Eval(ctx, evalArgs...)
	if err != nil {
		return report.Report{}, fmt.Errorf("error encountered in query evaluation %w", err)
	}

	result, err := resultSetToReport(resultSet, true)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to convert result set to report: %w", err)
	}

	for i := range result.Violations {
		result.Violations[i].IsAggregate = true
	}

	return result, nil
}

func resultSetToReport(resultSet rego.ResultSet, aggregate bool) (report.Report, error) {
	if len(resultSet) != 1 {
		return report.Report{}, fmt.Errorf("expected 1 item in resultset, got %d", len(resultSet))
	}

	r := report.Report{}

	if aggregate {
		if binding, ok := resultSet[0].Bindings["lint"].(map[string]any); ok {
			if aggregateBinding, ok := binding["aggregate"]; ok {
				if err := encoding.JSONRoundTrip(aggregateBinding, &r); err != nil {
					return report.Report{}, fmt.Errorf("JSON rountrip failed for bindings: %v %w", binding, err)
				}
			}
		}
	} else {
		if binding, ok := resultSet[0].Bindings["lint"]; ok {
			if err := encoding.JSONRoundTrip(binding, &r); err != nil {
				return report.Report{}, fmt.Errorf("JSON rountrip failed for bindings: %v %w", binding, err)
			}
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
		log.Println(outil.ByteSliceToString(bs))
	}

	l.combinedCfg = &mergedConf

	return l.combinedCfg, nil
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
