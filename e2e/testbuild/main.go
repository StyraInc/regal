package main

import (
	"github.com/open-policy-agent/opa/cmd"

	regal_cmd "github.com/open-policy-agent/regal/cmd"
)

// Create a frankenstein binary: it'll combine all of OPA and Regal.
// This way, we ensure that every Go module OPA could ever use, when
// running as server, can be used together with Regal's set of Go
// modules.
func main() {
	for _, c := range regal_cmd.RootCommand.Commands() {
		cmd.RootCommand.AddCommand(c)
	}
	if err := cmd.RootCommand.Execute(); err != nil {
		panic(err)
	}
}
