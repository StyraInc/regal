package main

import (
	"embed"
	"log"
	"os"

	"github.com/styrainc/regal/cmd"
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

	// The embedded FS can't point to parent directories, so while not pretty, we'll
	// need to pass it from here to the next command
	cmd.EmbedBundleFS = bundle

	if err := cmd.RootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
