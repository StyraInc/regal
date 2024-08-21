package cmd

import (
	"fmt"
	"os"

	"github.com/anderseknert/roast/pkg/encoding"
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
			bs, err := encoding.JSON().MarshalIndent(compile.Capabilities(), "", "  ")
			if err != nil {
				return fmt.Errorf("failed marshalling capabilities: %w", err)
			}

			fmt.Fprintln(os.Stdout, string(bs))

			return nil
		},
	}

	RootCommand.AddCommand(capabilitiesCommand)
}
