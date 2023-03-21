package cmd

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/open-policy-agent/opa/loader"
	"github.com/spf13/cobra"
	rio "github.com/styrainc/regal/internal/io"
	"github.com/styrainc/regal/pkg/config"
	"github.com/styrainc/regal/pkg/linter"
)

type lintCommandParams struct {
	timeout time.Duration
}

//nolint:gochecknoglobals
var EmbedBundleFS embed.FS

var errNoFileProvided = errors.New("at least one file or directory must be provided for linting")

func init() {
	params := lintCommandParams{}

	lintCommand := &cobra.Command{
		Use:   "lint <path> [path [...]]",
		Short: "Lint Rego source files",
		Long:  `Lint Rego source files for linter rule violations.`,

		PreRunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errNoFileProvided
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

	lintCommand.Flags().DurationVar(&params.timeout, "timeout", 0, "set timeout for linting (default unlimited)")

	RootCommand.AddCommand(lintCommand)
}

func lint(args []string, params lintCommandParams) error {
	ctx := context.Background()

	if params.timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, params.timeout)

		defer cancel()
	}

	// Create new fs from root of bundle, to avoid having to deal with
	// "bundle" in paths (i.e. `data.bundle.regal`)
	bfs, err := fs.Sub(EmbedBundleFS, "bundle")
	if err != nil {
		return fmt.Errorf("failed reading embedded bundle %w", err)
	}

	regalRules := rio.MustLoadRegalBundleFS(bfs)

	policies, err := loader.AllRegos(args)
	if err != nil {
		return fmt.Errorf("failed to load policy from provided args %w", err)
	}

	// TODO: Allow user-provided path to config
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get cwd %w", err)
	}

	regal := linter.NewLinter().WithAddedBundle(regalRules)

	userConfig, err := config.FindConfig(cwd)
	if err == nil {
		defer rio.CloseFileIgnore(userConfig)

		regal = regal.WithUserConfig(rio.MustYAMLToMap(userConfig))
	}

	rep, err := regal.Lint(ctx, policies)
	if err != nil {
		return fmt.Errorf("error(s) ecountered while linting %w", err)
	}

	// TODO: Create reporter interface and implementations
	log.Println(rep)

	return nil
}
