package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindConfigRoots will search for all config roots in the given path. A config
// root is a directory that contains a .regal.yaml file or a .regal/config.yaml
// file. This is intended to be used to by the language server when determining
// the most suitable workspace root for the server to operate on.
func FindConfigRoots(path string) ([]string, error) {
	var foundRoots []string

	err := filepath.WalkDir(path, func(path string, info os.DirEntry, _ error) error {
		if info.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		if filepath.Base(path) == ".regal.yaml" {
			foundRoots = append(foundRoots, filepath.Dir(path))
		}

		if strings.HasSuffix(path, ".regal/config.yaml") {
			foundRoots = append(foundRoots, filepath.Dir(filepath.Dir(path)))
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return foundRoots, nil
}
