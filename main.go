package main

import (
	"embed"
	"log"
	"os"

	"github.com/styrainc/regal/cmd"
	"github.com/styrainc/regal/internal/embeds"
)

// Note: this will bundle the tests as well, but since that has negligible impact on the size of the binary,
// it's preferable to filter them out from the bundle than to e.g. create a separate directory for tests
//
//go:embed all:bundle
var bundle embed.FS

func main() {
	// Remove date and time from any `log.*` calls, as that doesn't add much of value here
	// Evaluate options for logging later..
	log.SetFlags(0)

	// Embedded files and directories can only traverse down in the tree, i.e. no parent paths,
	// so while this isn't pretty, we'll use a separate package to carry state from here. If you
	// know of a better way of doing this, don't hesitate to fix this.
	embeds.EmbedBundleFS = bundle

	if err := cmd.RootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
