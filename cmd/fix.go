package cmd

import (
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

	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/fixer"
	"github.com/styrainc/regal/pkg/linter"
	"github.com/styrainc/regal/pkg/report"
)

// fixCommandParams is similar to the lint params, but with some fields such as profiling removed.
// It is intended that it is compatible with the same command line flags as the lint command to
// control the behavior of lint rules used.
type fixCommandParams struct {
	timeout         time.Duration
	configFile      string
	outputFile      string
	rules           repeatedStringFlag
	noColor         bool
	debug           bool
	disable         repeatedStringFlag
	disableAll      bool
	disableCategory repeatedStringFlag
	enable          repeatedStringFlag
	enableAll       bool
	enableCategory  repeatedStringFlag
	ignoreFiles     repeatedStringFlag
}

func (p *fixCommandParams) getConfigFile() string {
	return p.configFile
}

func (p *fixCommandParams) getTimeout() time.Duration {
	return p.timeout
}

func init() {
	params := &fixCommandParams{}

	fixCommand := &cobra.Command{
		Use:   "fix <path> [path [...]]",
		Short: "Fix Rego source files",
		Long:  `Fix Rego source files with linter violations.`,

		PreRunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("at least one file or directory must be provided for fixing")
			}

			return nil
		},

		RunE: wrapProfiling(func(args []string) error {
			err := fix(args, params)
			if err != nil {
				log.SetOutput(os.Stderr)
				log.Println(err)

				return exit(1)
			}

			return nil
		}),
	}

	fixCommand.Flags().StringVarP(&params.configFile, "config-file", "c", "",
		"set path of configuration file")
	fixCommand.Flags().StringVarP(&params.outputFile, "output-file", "o", "",
		"set file to use for fixing output, defaults to stdout")
	fixCommand.Flags().BoolVar(&params.noColor, "no-color", false,
		"Disable color output")
	fixCommand.Flags().VarP(&params.rules, "rules", "r",
		"set custom rules file(s). This flag can be repeated.")
	fixCommand.Flags().DurationVar(&params.timeout, "timeout", 0,
		"set timeout for fixing (default unlimited)")
	fixCommand.Flags().BoolVar(&params.debug, "debug", false,
		"enable debug logging (including print output from custom policy)")

	fixCommand.Flags().VarP(&params.disable, "disable", "d",
		"disable specific rule(s). This flag can be repeated.")
	fixCommand.Flags().BoolVarP(&params.disableAll, "disable-all", "D", false,
		"disable all rules")
	fixCommand.Flags().VarP(&params.disableCategory, "disable-category", "",
		"disable all rules in a category. This flag can be repeated.")

	fixCommand.Flags().VarP(&params.enable, "enable", "e",
		"enable specific rule(s). This flag can be repeated.")
	fixCommand.Flags().BoolVarP(&params.enableAll, "enable-all", "E", false,
		"enable all rules")
	fixCommand.Flags().VarP(&params.enableCategory, "enable-category", "",
		"enable all rules in a category. This flag can be repeated.")

	fixCommand.Flags().VarP(&params.ignoreFiles, "ignore-files", "",
		"ignore all files matching a glob-pattern. This flag can be repeated.")

	addPprofFlag(fixCommand.Flags())

	RootCommand.AddCommand(fixCommand)
}

func fix(args []string, params *fixCommandParams) error {
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
			return fmt.Errorf("failed to open output file before use %w", err)
		}
	}

	var regalDir *os.File

	var customRulesDir string

	var configSearchPath string

	cwd, _ := os.Getwd()

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

	regal := linter.NewLinter().
		WithDisableAll(params.disableAll).
		WithDisabledCategories(params.disableCategory.v...).
		WithDisabledRules(params.disable.v...).
		WithEnableAll(params.enableAll).
		WithEnabledCategories(params.enableCategory.v...).
		WithEnabledRules(params.enable.v...).
		WithDebugMode(params.debug).
		WithInputPaths(args)

	if customRulesDir != "" {
		regal = regal.WithCustomRules([]string{customRulesDir})
	}

	if params.rules.isSet {
		regal = regal.WithCustomRules(params.rules.v)
	}

	if params.ignoreFiles.isSet {
		regal = regal.WithIgnore(params.ignoreFiles.v)
	}

	var userConfig config.Config

	userConfigFile, err := readUserConfig(params, regalDir)

	switch {
	case err == nil:
		defer rio.CloseFileIgnore(userConfigFile)

		if params.debug {
			log.Printf("found user config file: %s", userConfigFile.Name())
		}

		if err := yaml.NewDecoder(userConfigFile).Decode(&userConfig); err != nil {
			if regalDir != nil {
				return fmt.Errorf("failed to decode user config from %s: %w", regalDir.Name(), err)
			}

			return fmt.Errorf("failed to decode user config: %w", err)
		}

		regal = regal.WithUserConfig(userConfig)
	case err != nil && params.configFile != "":
		return fmt.Errorf("user-provided config file not found: %w", err)
	case params.debug:
		log.Println("no user-provided config file found, will use the default config")
	}

	lintReport, err := regal.Lint(ctx)
	if err != nil {
		return fmt.Errorf("error(s) encountered while linting: %w", err)
	}

	fixReport, err := fixLintReport(&lintReport)
	if err != nil {
		return fmt.Errorf("error(s) encountered while fixing: %w", err)
	}

	fixer.Reporter(outputWriter, fixReport)

	return nil
}

func fixLintReport(rep *report.Report) (*fixer.Report, error) {
	fileReaders := make(map[string]io.Reader)

	for _, v := range rep.Violations {
		// if the file has already been opened, skip it
		if _, ok := fileReaders[v.Location.File]; ok {
			continue
		}

		f, err := os.Open(v.Location.File)
		if err != nil {
			return nil, fmt.Errorf("failed to open file for fixing %s: %w", v.Location.File, err)
		}

		defer f.Close()

		fileReaders[v.Location.File] = f
	}

	f := fixer.Fixer{}
	f.RegisterFixes(fixer.NewDefaultFixes()...)

	fixReport, err := f.Fix(rep, fileReaders)
	if err != nil {
		return nil, fmt.Errorf("error encountered while fixing: %w", err)
	}

	for file, content := range fixReport.FileContents() {
		stat, err := os.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("failed to get file info for file %s: %w", file, err)
		}

		f, err := os.OpenFile(file, os.O_RDWR|os.O_TRUNC, stat.Mode())
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", file, err)
		}

		_, err = f.Write(content)
		if err != nil {
			return nil, fmt.Errorf("failed to write to file %s: %w", file, err)
		}

		err = f.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close file %s: %w", file, err)
		}
	}

	return fixReport, nil
}
