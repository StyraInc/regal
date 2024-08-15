package bundles

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/open-policy-agent/opa/bundle"
)

// LoadDataBundle loads a bundle from the given path but only includes data
// files. The path must contain a bundle manifest file.
func LoadDataBundle(path string) (bundle.Bundle, error) {
	if _, err := os.Stat(filepath.Join(path, ".manifest")); err != nil {
		return bundle.Bundle{}, fmt.Errorf("manifest file was not found at bundle path %q", path)
	}

	b, err := bundle.NewCustomReader(bundle.NewDirectoryLoader(path).WithFilter(dataFileLoaderFilter)).Read()
	if err != nil {
		return bundle.Bundle{}, fmt.Errorf("failed to read bundle: %w", err)
	}

	return b, nil
}

func dataFileLoaderFilter(abspath string, info os.FileInfo, _ int) bool {
	if info.IsDir() {
		return false
	}

	basename := filepath.Base(abspath)

	return !slices.Contains([]string{".manifest", "data.json", "data.yml", "data.yaml"}, basename)
}
