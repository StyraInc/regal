//nolint:wrapcheck
package cmd

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-policy-agent/opa/ast"

	rp "github.com/styrainc/regal/internal/parse"
)

func init() {
	parseCommand := &cobra.Command{
		Use:   "parse <path> [path [...]]",
		Short: "Parse Rego source files with Regal enhancements included in output",
		Long:  "This command works similar to `opa parse` but includes Regal enhancements in the AST output.",

		PreRunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no file to parse provided")
			}

			if len(args) > 1 {
				return errors.New("only one file can be parsed at a time")
			}

			return nil
		},

		Run: func(_ *cobra.Command, args []string) {
			if err := parse(args); err != nil {
				log.SetOutput(os.Stderr)
				log.Println(err)
				os.Exit(1)
			}
		},
	}
	RootCommand.AddCommand(parseCommand)
}

func parse(args []string) error {
	filename := args[0]

	bs, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	content := string(bs)

	module, err := ast.ParseModuleWithOpts(filename, content, rp.ParserOptions())
	if err != nil {
		return err
	}

	enhancedAST, err := rp.PrepareAST(filename, content, module)
	if err != nil {
		return err
	}

	bs, err = json.MarshalIndent(enhancedAST, "", "  ")
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(bs)

	return err
}
