package cmd

import (
	"os"
	"path"

	"github.com/spf13/cobra"
)

// RootCommand is the base CLI command that all subcommands are added to.
//
//nolint:gochecknoglobals
var RootCommand = &cobra.Command{
	Use:   path.Base(os.Args[0]),
	Short: "Regal",
	Long:  "Regal is a linter for Rego, with the goal of making your Rego magnificent!",
}
