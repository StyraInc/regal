package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/metrics"
	"github.com/open-policy-agent/opa/topdown"

	rio "github.com/styrainc/regal/internal/io"
	regalmetrics "github.com/styrainc/regal/internal/metrics"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/linter"
	"github.com/styrainc/regal/pkg/report"
	"github.com/styrainc/regal/pkg/reporter"
)

type lintCommandParams struct {
	timeout         time.Duration
	configFile      string
	format          string
	outputFile      string
	failLevel       string
	rules           repeatedStringFlag
	noColor         bool
	debug           bool
	enablePrint     bool
	metrics         bool
	profile         bool
	disable         repeatedStringFlag
	disableAll      bool
	disableCategory repeatedStringFlag
	enable          repeatedStringFlag
	enableAll       bool
	enableCategory  repeatedStringFlag
	ignoreFiles     repeatedStringFlag
}

const stringType = "string"

type repeatedStringFlag struct {
	v     []string
	isSet bool
}

func (f *repeatedStringFlag) Type() string {
	return stringType
}

func (f *repeatedStringFlag) String() string {
	return strings.Join(f.v, ",")
}

func (f *repeatedStringFlag) Set(s string) error {
	f.v = append(f.v, s)
	f.isSet = true

	return nil
}

func init() {
	params := lintCommandParams{}

	lintCommand := &cobra.Command{
		Use:   "lint <path> [path [...]]",
		Short: "Lint Rego source files",
		Long:  `Lint Rego source files for linter rule violations.`,

		PreRunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("at least one file or directory must be provided for linting")
			}

			return nil
		},

		RunE: wrapProfiling(func(args []string) error {
			// Allow setting debug mode via GitHub UI for failing actions
			if os.Getenv("RUNNER_DEBUG") != "" {
				params.debug = true
			}

			rep, err := lint(args, params)
			if err != nil {
				log.SetOutput(os.Stderr)
				log.Println(err)

				return exit(1)
			}

			errorsFound := 0
			warningsFound := 0

			for _, violation := range rep.Violations {
				if violation.Level == "error" {
					errorsFound++
				} else if violation.Level == "warning" {
					warningsFound++
				}
			}

			exitCode := 0

			if params.failLevel == "error" && errorsFound > 0 {
				exitCode = 3
			}

			if params.failLevel == "warning" {
				if errorsFound > 0 {
					exitCode = 3
				} else if warningsFound > 0 {
					exitCode = 2
				}
			}
			if exitCode != 0 {
				return exit(exitCode)
			}

			return nil
		}),
	}

	lintCommand.Flags().StringVarP(&params.configFile, "config-file", "c", "",
		"set path of configuration file")
	lintCommand.Flags().StringVarP(&params.format, "format", "f", formatPretty,
		"set output format (pretty, compact, json, github)")
	lintCommand.Flags().StringVarP(&params.outputFile, "output-file", "o", "",
		"set file to use for linting output, defaults to stdout")
	lintCommand.Flags().StringVarP(&params.failLevel, "fail-level", "l", "error",
		"set level at which to fail with a non-zero exit code (error, warning)")
	lintCommand.Flags().BoolVar(&params.noColor, "no-color", false,
		"Disable color output")
	lintCommand.Flags().VarP(&params.rules, "rules", "r",
		"set custom rules file(s). This flag can be repeated.")
	lintCommand.Flags().DurationVar(&params.timeout, "timeout", 0,
		"set timeout for linting (default unlimited)")
	lintCommand.Flags().BoolVar(&params.debug, "debug", false,
		"enable debug logging (including print output from custom policy)")
	lintCommand.Flags().BoolVar(&params.enablePrint, "enable-print", false,
		"enable print output from policy")
	lintCommand.Flags().BoolVar(&params.metrics, "metrics", false,
		"enable metrics reporting (currently supported only for JSON output format)")
	lintCommand.Flags().BoolVar(&params.profile, "profile", false,
		"enable profiling metrics to be added to reporting (currently supported only for JSON output format)")

	lintCommand.Flags().VarP(&params.disable, "disable", "d",
		"disable specific rule(s). This flag can be repeated.")
	lintCommand.Flags().BoolVarP(&params.disableAll, "disable-all", "D", false,
		"disable all rules")
	lintCommand.Flags().VarP(&params.disableCategory, "disable-category", "",
		"disable all rules in a category. This flag can be repeated.")

	lintCommand.Flags().VarP(&params.enable, "enable", "e",
		"enable specific rule(s). This flag can be repeated.")
	lintCommand.Flags().BoolVarP(&params.enableAll, "enable-all", "E", false,
		"enable all rules")
	lintCommand.Flags().VarP(&params.enableCategory, "enable-category", "",
		"enable all rules in a category. This flag can be repeated.")

	lintCommand.Flags().VarP(&params.ignoreFiles, "ignore-files", "",
		"ignore all files matching a glob-pattern. This flag can be repeated.")

	addPprofFlag(lintCommand.Flags())

	RootCommand.AddCommand(lintCommand)
}

//nolint:gocognit
func lint(args []string, params lintCommandParams) (report.Report, error) {
	var err error

	ctx, cancel := getLinterContext(params)
	defer cancel()

	if params.noColor {
		color.NoColor = true
	}

	// if an outputFile has been set, open it for writing or create it
	var outputWriter io.Writer

	outputWriter = os.Stdout
	if params.outputFile != "" {
		outputWriter, err = getWriterForOutputFile(params.outputFile)
		if err != nil {
			return report.Report{}, fmt.Errorf("failed to open output file before use %w", err)
		}
	}

	var regalDir *os.File

	var customRulesDir string

	var configSearchPath string

	cwd, _ := os.Getwd()

	m := metrics.New()
	if params.metrics {
		m.Timer(regalmetrics.RegalConfigSearch).Start()
	}

	if len(args) == 1 {
		configSearchPath = args[0]
		if !strings.HasPrefix(args[0], "/") {
			configSearchPath = filepath.Join(cwd, args[0])
		}
	} else {
		configSearchPath, _ = os.Getwd()
	}

	if configSearchPath == "" {
		log.Println("failed to determine relevant directory for config file search - won't search for custom config or rules")
	} else {
		regalDir, err = config.FindRegalDirectory(configSearchPath)
		if err == nil {
			customRulesPath := filepath.Join(regalDir.Name(), rio.PathSeparator, "rules")
			if _, err = os.Stat(customRulesPath); err == nil {
				customRulesDir = customRulesPath
			}
		}
	}

	if params.metrics {
		m.Timer(regalmetrics.RegalConfigSearch).Stop()
	}

	regal := linter.NewLinter().
		WithDisableAll(params.disableAll).
		WithDisabledCategories(params.disableCategory.v...).
		WithDisabledRules(params.disable.v...).
		WithEnableAll(params.enableAll).
		WithEnabledCategories(params.enableCategory.v...).
		WithEnabledRules(params.enable.v...).
		WithDebugMode(params.debug).
		WithInputPaths(args)

	if params.enablePrint {
		regal = regal.WithPrintHook(topdown.NewPrintHook(os.Stderr))
	}

	if customRulesDir != "" {
		regal = regal.WithCustomRules([]string{customRulesDir})
	}

	if params.rules.isSet {
		regal = regal.WithCustomRules(params.rules.v)
	}

	if params.ignoreFiles.isSet {
		regal = regal.WithIgnore(params.ignoreFiles.v)
	}

	if params.metrics {
		regal = regal.WithMetrics(m)
		m.Timer(regalmetrics.RegalConfigParse).Start()
	}

	if params.profile {
		regal = regal.WithProfiling(true)
	}

	var userConfig config.Config

	userConfigFile, err := readUserConfig(params, regalDir)
	if err == nil { //nolint:nestif
		defer rio.CloseFileIgnore(userConfigFile)

		if params.debug {
			log.Printf("found user config file: %s", userConfigFile.Name())
		}

		if err := yaml.NewDecoder(userConfigFile).Decode(&userConfig); err != nil {
			if regalDir != nil {
				return report.Report{}, fmt.Errorf("failed to decode user config from %s: %w", regalDir.Name(), err)
			}

			return report.Report{}, fmt.Errorf("failed to decode user config: %w", err)
		}

		regal = regal.WithUserConfig(userConfig)
	} else if params.debug {
		log.Println("no user-provided config file found, will use the default config")
	}

	if params.metrics {
		m.Timer(regalmetrics.RegalConfigParse).Stop()
	}

	result, err := regal.Lint(ctx)
	if err != nil {
		return report.Report{}, fmt.Errorf("error(s) encountered while linting: %w", err)
	}

	rep, err := getReporter(params.format, outputWriter)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to get reporter: %w", err)
	}

	return result, rep.Publish(ctx, result) //nolint:wrapcheck
}

func getReporter(format string, outputWriter io.Writer) (reporter.Reporter, error) {
	switch format {
	case formatPretty:
		return reporter.NewPrettyReporter(outputWriter), nil
	case formatCompact:
		return reporter.NewCompactReporter(outputWriter), nil
	case formatJSON:
		return reporter.NewJSONReporter(outputWriter), nil
	case formatGitHub:
		return reporter.NewGitHubReporter(outputWriter), nil
	default:
		return nil, fmt.Errorf("unknown format %s", format)
	}
}

func readUserConfig(params lintCommandParams, regalDir *os.File) (userConfig *os.File, err error) {
	if params.configFile != "" {
		userConfig, err = os.Open(params.configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file %w", err)
		}
	} else {
		searchPath, _ := os.Getwd()
		if regalDir != nil {
			searchPath = regalDir.Name()
		}
		if searchPath != "" {
			userConfig, err = config.FindConfig(searchPath)
		}
	}

	return userConfig, err //nolint:wrapcheck
}

func getLinterContext(params lintCommandParams) (context.Context, func()) {
	ctx := context.Background()

	cancel := func() {}

	if params.timeout != 0 {
		ctx, cancel = context.WithTimeout(ctx, params.timeout)
	}

	return ctx, cancel
}

func getWriterForOutputFile(filename string) (io.Writer, error) {
	if _, err := os.Stat(filename); err == nil {
		f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755)
		if err != nil {
			return nil, fmt.Errorf("failed to open output file %w", err)
		}

		return f, nil
	}

	f, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file %w", err)
	}

	return f, nil
}
