package main

import (
	"log"
	"os"

	"github.com/styrainc/regal/cmd"
)

func main() {
	// Remove date and time from any `log.*` calls, as that doesn't add much of value here
	// Evaluate options for logging later
	log.SetFlags(0)

	// TODO: Temporary until this requirement is removed from OPA
	if err := os.Setenv("EXPERIMENTAL_GENERAL_RULE_REFS", "true"); err != nil {
		log.Fatal(err)
	}

	if err := cmd.RootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
