package config

import (
	"path/filepath"
	"strings"

	"github.com/styrainc/regal/internal/io/files"
	"github.com/styrainc/regal/internal/io/files/filter"
)

// FindConfigRoots will search for all config roots in the given path. A config
// root is a directory that contains a .regal.yaml file or a .regal/config.yaml
// file. This is intended to be used to by the language server when determining
// the most suitable config root for the server to operate on.
func FindConfigRoots(path string) ([]string, error) {
	return files.DefaultWalkReducer(path, []string{}).
		WithFilters(filter.Not(filter.Filenames("config.yaml", ".regal.yaml"))).
		Reduce(configRootsReducer)
}

func configRootsReducer(path string, foundRoots []string) ([]string, error) {
	if strings.HasSuffix(path, ".regal.yaml") {
		foundRoots = append(foundRoots, filepath.Dir(path))
	}

	if strings.HasSuffix(path, ".regal/config.yaml") {
		foundRoots = append(foundRoots, filepath.Dir(filepath.Dir(path)))
	}

	return foundRoots, nil
}
