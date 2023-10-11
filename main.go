package main

import (
	"errors"
	"log"
	"os"

	"github.com/styrainc/regal/cmd"
)

func main() {
	// Remove date and time from any `log.*` calls, as that doesn't add much of value here
	// Evaluate options for logging later
	log.SetFlags(0)

	if err := cmd.RootCommand.Execute(); err != nil {
		code := 1
		if e := (cmd.ExitError{}); errors.As(err, &e) {
			code = e.Code()
		}

		os.Exit(code)
	}
}
