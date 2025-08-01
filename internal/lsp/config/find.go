package config

import (
	"path/filepath"
	"strings"

	rio "github.com/styrainc/regal/internal/io"
)

// FindConfigRoots will search for all config roots in the given path. A config
// root is a directory that contains a .regal.yaml file or a .regal/config.yaml
// file. This is intended to be used to by the language server when determining
// the most suitable config root for the server to operate on.
func FindConfigRoots(path string) ([]string, error) {
	var foundRoots []string

	return rio.NewFileWalkReducer(path, foundRoots).
		WithFilters(rio.DirectoryFilter, rio.NegateFilter(rio.SuffixesFilter(".yaml"))).
		WithSkipFunc(rio.DefaultSkipDirectories).
		Reduce(configRootsReducer)
}

func configRootsReducer(path string, foundRoots []string) ([]string, error) {
	if filepath.Base(path) == ".regal.yaml" {
		foundRoots = append(foundRoots, filepath.Dir(path))
	}

	if strings.HasSuffix(path, ".regal/config.yaml") {
		foundRoots = append(foundRoots, filepath.Dir(filepath.Dir(path)))
	}

	return foundRoots, nil
}
