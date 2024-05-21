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
	"github.com/styrainc/regal/pkg/fixer/fileprovider"
	"github.com/styrainc/regal/pkg/fixer/fixes"
	"github.com/styrainc/regal/pkg/linter"
)

// fixCommandParams is similar to the lint params, but with some fields such as profiling removed.
// It is intended that it is compatible with the same command line flags as the lint command to
// control the behavior of lint rules used.
type fixCommandParams struct {
	configFile      string
	debug           bool
	disable         repeatedStringFlag
	disableAll      bool
	disableCategory repeatedStringFlag
	enable          repeatedStringFlag
	enableAll       bool
	enableCategory  repeatedStringFlag
	format          string
	ignoreFiles     repeatedStringFlag
	noColor         bool
	outputFile      string
	rules           repeatedStringFlag
	timeout         time.Duration
}

func (p *fixCommandParams) getConfigFile() string {
	return p.configFile
}

func (p *fixCommandParams) getTimeout() time.Duration {
	return p.timeout
}

func init() {
	params := &fixCommandParams{}

	intro := strings.ReplaceAll(`Fix Rego source files with linter violations.
Note that this command is intended to help fix style-related issues,
and could be considered as a stricter opa-fmt command.
Issues like bugs should be fixed manually,
as it is important to understand why the were flagged.`, "\n", " ")

	fixCommand := &cobra.Command{
		Use:   "fix <path> [path [...]]",
		Short: "Fix Rego source files",
		Long: func() string {
			var fixableRules []string
			for _, f := range fixes.NewDefaultFixes() {
				fixableRules = append(fixableRules, f.Name())
			}

			if len(fixableRules) == 0 {
				return intro
			}

			return fmt.Sprintf(`%s

The linter rules with automatic fixes available are currently:
- %s
`, intro, strings.Join(fixableRules, "\n- "))
		}(),
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
	fixCommand.Flags().StringVarP(&params.format, "format", "f", formatPretty,
		"set output format (pretty is the only supported format)")
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

	l := linter.NewLinter().
		WithDisableAll(params.disableAll).
		WithDisabledCategories(params.disableCategory.v...).
		WithDisabledRules(params.disable.v...).
		WithEnableAll(params.enableAll).
		WithEnabledCategories(params.enableCategory.v...).
		WithEnabledRules(params.enable.v...).
		WithDebugMode(params.debug)

	if customRulesDir != "" {
		l = l.WithCustomRules([]string{customRulesDir})
	}

	if params.rules.isSet {
		l = l.WithCustomRules(params.rules.v)
	}

	if params.ignoreFiles.isSet {
		l = l.WithIgnore(params.ignoreFiles.v)
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

		l = l.WithUserConfig(userConfig)
	case err != nil && params.configFile != "":
		return fmt.Errorf("user-provided config file not found: %w", err)
	case params.debug:
		log.Println("no user-provided config file found, will use the default config")
	}

	f := fixer.NewFixer()
	f.RegisterFixes(fixes.NewDefaultFixes()...)

	fileProvider := fileprovider.NewFSFileProvider(args...)

	fixReport, err := f.Fix(ctx, &l, fileProvider)
	if err != nil {
		return fmt.Errorf("failed to fix: %w", err)
	}

	r, err := fixer.ReporterForFormat(params.format, outputWriter)
	if err != nil {
		return fmt.Errorf("failed to create reporter for format %s: %w", params.format, err)
	}

	err = r.Report(fixReport)
	if err != nil {
		return fmt.Errorf("failed to output fix report: %w", err)
	}

	return nil
}
