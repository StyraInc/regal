package config

import (
	"os"
	"path/filepath"
)

// GlobalDir is the config directory that will be used for user-wide configuration.
// This is different from the .regal directories that are searched for when
// linting.
func GlobalDir() string {
	cfgDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	regalDir := filepath.Join(cfgDir, ".config", "regal")
	if _, err := os.Stat(regalDir); os.IsNotExist(err) {
		if err := os.Mkdir(regalDir, os.ModePerm); err != nil {
			return ""
		}
	}

	return regalDir
}
