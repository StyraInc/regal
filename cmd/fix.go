package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"

	"github.com/styrainc/regal/internal/git"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/internal/util"
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
	dryRun          bool
	verbose         bool
	force           bool
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
and could be considered as a stricter opa fmt command.
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

	fixCommand.Flags().BoolVarP(&params.dryRun, "dry-run", "", false,
		"run the fixer in dry-run mode, use with --verbose to see changes")

	fixCommand.Flags().BoolVarP(&params.verbose, "verbose", "", false,
		"show the full changes applied in the console")

	fixCommand.Flags().BoolVarP(&params.force, "force", "", false,
		"allow fixing of files that have uncommitted changes in git or when git is not being used")

	addPprofFlag(fixCommand.Flags())

	RootCommand.AddCommand(fixCommand)
}

// TODO: This function is too long and should be broken down
//
//nolint:maintidx
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

		err := yaml.NewDecoder(userConfigFile).Decode(&userConfig)
		if errors.Is(err, io.EOF) {
			log.Printf("user config file %q is empty, will use the default config", userConfigFile.Name())
		} else if err != nil {
			if regalDir != nil {
				return fmt.Errorf("failed to decode user config from %s: %w", regalDir.Name(), err)
			}

			return fmt.Errorf("failed to decode user config: %w", err)
		}

		l = l.WithUserConfig(userConfig)
	case params.configFile != "":
		return fmt.Errorf("user-provided config file not found: %w", err)
	case params.debug:
		log.Println("no user-provided config file found, will use the default config")
	}

	roots, err := config.GetPotentialRoots(args...)
	if err != nil {
		return fmt.Errorf("could not find potential roots: %w", err)
	}

	f := fixer.NewFixer()
	f.RegisterRoots(roots...)
	f.RegisterFixes(fixes.NewDefaultFixes()...)
	f.RegisterMandatoryFixes(
		&fixes.Fmt{
			NameOverride: "use-rego-v1",
			OPAFmtOpts: format.Opts{
				RegoVersion: ast.RegoV0CompatV1,
			},
		},
	)

	ignore := userConfig.Ignore.Files

	if len(params.ignoreFiles.v) > 0 {
		ignore = params.ignoreFiles.v
	}

	// create a list of absolute paths, these will be used for the file from
	// this point in order to be able to use the roots for format reporting.
	absArgs := make([]string, len(args))

	for i, arg := range args {
		if filepath.IsAbs(arg) {
			absArgs[i] = arg

			continue
		}

		absArgs[i], err = filepath.Abs(arg)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", arg, err)
		}
	}

	filtered, err := config.FilterIgnoredPaths(args, ignore, true, "")
	if err != nil {
		return fmt.Errorf("failed to filter ignored paths: %w", err)
	}

	slices.Sort(filtered)
	// TODO: Figure out why filtered returns duplicates in the first place
	filtered = slices.Compact(filtered)

	fileProvider, err := fileprovider.NewInMemoryFileProviderFromFS(filtered...)
	if err != nil {
		return fmt.Errorf("failed to create file provider: %w", err)
	}

	fixReport, err := f.Fix(ctx, &l, fileProvider)
	if err != nil {
		return fmt.Errorf("failed to fix: %w", err)
	}

	gitRepo, err := git.FindGitRepo(args...)
	if err != nil {
		return fmt.Errorf("failed to establish git repo: %w", err)
	}

	if gitRepo == "" && !params.force {
		return errors.New("no git repo found to support undo, use --force to override")
	}

	// if the fixer is being run in a git repo, we must not fix files that have
	// been changed.
	if !params.dryRun && !params.force {
		changedFiles := make(map[string]struct{})

		if gitRepo != "" {
			cf, err := git.GetChangedFiles(gitRepo)
			if err != nil {
				return fmt.Errorf("failed to get changed files: %w", err)
			}

			for _, f := range cf {
				changedFiles[f] = struct{}{}
			}
		}

		var conflictingFiles []string

		for _, file := range fileProvider.ModifiedFiles() {
			if _, ok := changedFiles[file]; ok {
				conflictingFiles = append(conflictingFiles, file)
			}
		}

		if len(conflictingFiles) > 0 {
			return fmt.Errorf(
				`the following files have been changed since the fixer was run:
- %s
please run fix from a clean state to support the use of git checkout for undo`,
				strings.Join(conflictingFiles, "\n- "),
			)
		}
	}

	if params.verbose {
		if params.dryRun {
			fmt.Fprintln(outputWriter, "Dry run mode enabled, the following changes would be made:")
		}

		for _, file := range fileProvider.ModifiedFiles() {
			fmt.Fprintln(outputWriter, "Set:", file, "to:")

			fc, err := fileProvider.Get(file)
			if err != nil {
				return fmt.Errorf("failed to get file %s: %w", file, err)
			}

			fmt.Fprintln(outputWriter, string(fc))
			fmt.Fprintln(outputWriter, "----------")
		}

		for _, file := range fileProvider.DeletedFiles() {
			fmt.Fprintln(outputWriter, "Delete:", file)
		}
	}

	if !params.dryRun {
		for _, file := range fileProvider.DeletedFiles() {
			err := os.Remove(file)
			if err != nil {
				return fmt.Errorf("failed to delete file %s: %w", file, err)
			}

			err = util.DeleteEmptyDirs(filepath.Dir(file))
			if err != nil {
				return fmt.Errorf("failed to delete empty directories: %w", err)
			}
		}

		for _, file := range fileProvider.ModifiedFiles() {
			fc, err := fileProvider.Get(file)
			if err != nil {
				return fmt.Errorf("failed to get file %s: %w", file, err)
			}

			fileMode := fs.FileMode(0o600)

			fileInfo, err := os.Stat(file)
			if err == nil {
				fileMode = fileInfo.Mode()
			}

			err = os.MkdirAll(filepath.Dir(file), 0o755)
			if err != nil {
				return fmt.Errorf("failed to create directory for file %s: %w", file, err)
			}

			err = os.WriteFile(file, fc, fileMode)
			if err != nil {
				return fmt.Errorf("failed to write file %s: %w", file, err)
			}
		}
	}

	r, err := fixer.ReporterForFormat(params.format, outputWriter)
	if err != nil {
		return fmt.Errorf("failed to create reporter for format %s: %w", params.format, err)
	}

	r.SetDryRun(params.dryRun)

	err = r.Report(fixReport)
	if err != nil {
		return fmt.Errorf("failed to output fix report: %w", err)
	}

	return nil
}
