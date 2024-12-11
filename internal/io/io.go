package io

import (
	"fmt"
	"io"
	files "io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anderseknert/roast/pkg/encoding"
	"gopkg.in/yaml.v3"

	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/loader/filter"
)

const PathSeparator = string(os.PathSeparator)

// LoadRegalBundleFS loads bundle embedded from policy and data directory.
func LoadRegalBundleFS(fs files.FS) (bundle.Bundle, error) {
	embedLoader, err := bundle.NewFSLoader(fs)
	if err != nil {
		return bundle.Bundle{}, fmt.Errorf("failed to load bundle from filesystem: %w", err)
	}

	//nolint:wrapcheck
	return bundle.NewCustomReader(embedLoader.WithFilter(ExcludeTestFilter())).
		WithSkipBundleVerification(true).
		WithProcessAnnotations(true).
		WithBundleName("regal").
		Read()
}

// LoadRegalBundlePath loads bundle from path.
func LoadRegalBundlePath(path string) (bundle.Bundle, error) {
	//nolint:wrapcheck
	return bundle.NewCustomReader(bundle.NewDirectoryLoader(path).WithFilter(ExcludeTestFilter())).
		WithSkipBundleVerification(true).
		WithProcessAnnotations(true).
		WithBundleName("regal").
		Read()
}

// MustLoadRegalBundleFS loads bundle embedded from policy and data directory, exit on failure.
func MustLoadRegalBundleFS(fs files.FS) bundle.Bundle {
	regalBundle, err := LoadRegalBundleFS(fs)
	if err != nil {
		log.Fatal(err)
	}

	return regalBundle
}

// ToMap convert any value to map[string]any, or panics on failure.
func ToMap(a any) map[string]any {
	r := make(map[string]any)

	encoding.MustJSONRoundTrip(a, &r)

	return r
}

// CloseFileIgnore closes file ignoring errors, mainly for deferred cleanup.
func CloseFileIgnore(file *os.File) {
	_ = file.Close()
}

func ExcludeTestFilter() filter.LoaderFilter {
	return func(_ string, info files.FileInfo, _ int) bool {
		return strings.HasSuffix(info.Name(), "_test.rego") &&
			// (anderseknert): This is an outlier, but not sure we need something
			// more polished to deal with this for the time being.
			info.Name() != "todo_test.rego"
	}
}

// FindInput finds input.json or input.yaml file in workspace closest to the file, and returns
// both the location and the contents of the file (as map), or an empty string and nil if not found.
// Note that:
// - This function doesn't do error handling. If the file can't be read, nothing is returned.
// - While the input data theoritcally could be anything JSON/YAML value, we only support an object.
func FindInput(file string, workspacePath string) (string, map[string]any) {
	relative := strings.TrimPrefix(file, workspacePath)
	components := strings.Split(filepath.Dir(relative), PathSeparator)

	var (
		inputPath string
		content   []byte
	)

	for i := range components {
		current := components[:len(components)-i]

		inputPathJSON := filepath.Join(workspacePath, filepath.Join(current...), "input.json")

		f, err := os.Open(inputPathJSON)
		if err == nil {
			inputPath = inputPathJSON
			content, _ = io.ReadAll(f)

			break
		}

		inputPathYAML := filepath.Join(workspacePath, filepath.Join(current...), "input.yaml")

		f, err = os.Open(inputPathYAML)
		if err == nil {
			inputPath = inputPathYAML
			content, _ = io.ReadAll(f)

			break
		}
	}

	if inputPath == "" || content == nil {
		return "", nil
	}

	var input map[string]any

	if strings.HasSuffix(inputPath, ".json") {
		if err := encoding.JSON().Unmarshal(content, &input); err != nil {
			return "", nil
		}
	} else if strings.HasSuffix(inputPath, ".yaml") {
		if err := yaml.Unmarshal(content, &input); err != nil {
			return "", nil
		}
	}

	return inputPath, input
}

func IsSkipWalkDirectory(info files.DirEntry) bool {
	return info.IsDir() && (info.Name() == ".git" || info.Name() == ".idea" || info.Name() == "node_modules")
}

// WalkFiles walks the file system rooted at root, calling f for each file. This is
// a less ceremonious version of filepath.WalkDir where only file paths (not dirs)
// are passed to the callback, and where directories that should commonly  be ignored
// (.git, node_modules, etc.) are skipped.
func WalkFiles(root string, f func(path string) error) error {
	return filepath.WalkDir(root, func(path string, info os.DirEntry, _ error) error { //nolint:wrapcheck
		if IsSkipWalkDirectory(info) {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		return f(path)
	})
}

// FindManifestLocations walks the file system rooted at root, and returns the
// *relative* paths of directories containing a .manifest file.
func FindManifestLocations(root string) ([]string, error) {
	var foundBundleRoots []string

	if err := WalkFiles(root, func(path string) error {
		if filepath.Base(path) == ".manifest" {
			if rel, err := filepath.Rel(root, path); err == nil {
				foundBundleRoots = append(foundBundleRoots, filepath.Dir(rel))
			}
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk workspace path: %w", err)
	}

	return foundBundleRoots, nil
}
