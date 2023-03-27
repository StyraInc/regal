package cmd

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/linter"
	"github.com/styrainc/regal/pkg/rules"
)

type lintCommandParams struct {
	timeout    time.Duration
	configFile string
	rules      repeatedStringFlag
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

var EmbedBundleFS embed.FS //nolint:gochecknoglobals

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

		Run: func(_ *cobra.Command, args []string) {
			if err := lint(args, params); err != nil {
				log.SetOutput(os.Stderr)
				log.Println(err)
				os.Exit(1)
			}
		},
	}

	lintCommand.Flags().StringVarP(&params.configFile, "config-file", "c", "", "set path of configuration file")
	lintCommand.Flags().VarP(&params.rules, "rules", "r", "set custom rules file(s). This flag can be repeated.")
	lintCommand.Flags().DurationVar(&params.timeout, "timeout", 0, "set timeout for linting (default unlimited)")

	RootCommand.AddCommand(lintCommand)
}

func lint(args []string, params lintCommandParams) error {
	ctx, cancel := getLinterContext(params)
	defer cancel()

	// Create new fs from root of bundle, to avoid having to deal with
	// "bundle" in paths (i.e. `data.bundle.regal`)
	bfs, err := fs.Sub(EmbedBundleFS, "bundle")
	if err != nil {
		return fmt.Errorf("failed reading embedded bundle %w", err)
	}

	regalRules := rio.MustLoadRegalBundleFS(bfs)

	var regalDir *os.File

	var customRulesDir string

	cwd, err := os.Getwd()
	if err != nil {
		log.Println("failed to get current directory - won't search for custom config or rules")
	} else {
		regalDir, err = config.FindRegalDirectory(cwd)
		if err == nil {
			customRulesDir = filepath.Join(regalDir.Name(), rio.PathSeparator, "rules")
		}
	}

	regal := linter.NewLinter().WithAddedBundle(regalRules)

	if customRulesDir != "" {
		regal = regal.WithCustomRules([]string{customRulesDir})
	}

	if params.rules.isSet {
		regal = regal.WithCustomRules(params.rules.v)
	}

	userConfig, err := readUserConfig(params, regalDir)
	if err == nil {
		defer rio.CloseFileIgnore(userConfig)
		regal = regal.WithUserConfig(rio.MustYAMLToMap(userConfig))
	}

	input, err := rules.InputFromPaths(args)
	if err != nil {
		return fmt.Errorf("errors encountered when reading files to lint: %w", err)
	}

	rep, err := regal.Lint(ctx, input)
	if err != nil {
		return fmt.Errorf("error(s) ecountered while linting: %w", err)
	}

	// TODO: Create reporter interface and implementations
	log.Println(rep)

	return nil
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
