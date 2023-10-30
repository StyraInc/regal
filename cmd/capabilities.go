package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/styrainc/regal/internal/compile"
)

func init() {
	capabilitiesCommand := &cobra.Command{
		Hidden: true,
		Use:    "capabilities",
		Short:  "Print the capabilities of Regal",
		Long:   "Show capabilities for Regal",
		RunE: func(*cobra.Command, []string) error {
			bs, err := json.MarshalIndent(compile.DefaultCapabilities(), "", "  ")
			if err != nil {
				return fmt.Errorf("failed marshalling capabilities: %w", err)
			}

			fmt.Fprintln(os.Stdout, string(bs))

			return nil
		},
	}

	RootCommand.AddCommand(capabilitiesCommand)
}
