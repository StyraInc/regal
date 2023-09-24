package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/styrainc/regal/cmd"
)

func main() {
	// Remove date and time from any `log.*` calls, as that doesn't add much of value here
	// Evaluate options for logging later..
	log.SetFlags(0)

	// Get the full path of the executable file 'regal'
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}

	executableDir := filepath.Dir(executablePath)
	os.Setenv("EXECUTABLE_PATH", executableDir)

	if err := cmd.RootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
