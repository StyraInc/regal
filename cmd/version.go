package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/anderseknert/roast/pkg/encoding"
	"github.com/spf13/cobra"

	"github.com/styrainc/regal/pkg/version"
)

type versionCommandParams struct {
	format string
}

func init() {
	params := versionCommandParams{}

	parseCommand := &cobra.Command{
		Use:   "version [--format=json|pretty]",
		Short: "Print the version of Regal",
		Long:  "Show the version and other build-time metadata for the running Regal binary.",

		PreRunE: func(*cobra.Command, []string) error {
			if params.format == "" {
				params.format = "pretty"
			} else if params.format != "json" && params.format != "pretty" {
				return fmt.Errorf("invalid format: %s", params.format)
			}

			return nil
		},

		Run: func(*cobra.Command, []string) {
			vi := version.New()

			switch params.format {
			case formatJSON:
				e := encoding.JSON().NewEncoder(os.Stdout)
				e.SetIndent("", "  ")
				if err := e.Encode(vi); err != nil {
					log.SetOutput(os.Stderr)
					log.Println(err)
					os.Exit(1)
				}
			case formatPretty:
				os.Stdout.WriteString(vi.String())
			default:
				log.SetOutput(os.Stderr)
				log.Printf("invalid format: %s\n", params.format)
				os.Exit(1)
			}
		},
	}
	parseCommand.Flags().StringVar(
		&params.format,
		"format",
		formatPretty,
		fmt.Sprintf("Output format. Valid values are '%s' and '%s'.", formatPretty, formatJSON),
	)
	RootCommand.AddCommand(parseCommand)
}
