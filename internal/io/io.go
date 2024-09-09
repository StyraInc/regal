package io

import (
	"fmt"
	"io"
	files "io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/anderseknert/roast/pkg/encoding"

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

	MustJSONRoundTrip(a, &r)

	return r
}

// JSONRoundTrip convert any value to JSON and back again.
func JSONRoundTrip(from any, to any) error {
	json := encoding.JSON()

	bs, err := json.Marshal(from)
	if err != nil {
		return fmt.Errorf("failed JSON marshalling %w", err)
	}

	return json.Unmarshal(bs, to) //nolint:wrapcheck
}

// MustJSONRoundTrip convert any value to JSON and back again, exit on failure.
func MustJSONRoundTrip(from any, to any) {
	if err := JSONRoundTrip(from, to); err != nil {
		log.Fatal(err)
	}
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

// FindInput finds input.json file in workspace closest to the file, and returns
// both the location and the reader.
func FindInput(file string, workspacePath string) (string, io.Reader) {
	relative := strings.TrimPrefix(file, workspacePath)
	components := strings.Split(path.Dir(relative), string(filepath.Separator))

	for i := range len(components) {
		inputPath := path.Join(workspacePath, path.Join(components[:len(components)-i]...), "input.json")

		f, err := os.Open(inputPath)
		if err == nil {
			return inputPath, f
		}
	}

	return "", nil
}

func IsSkipWalkDirectory(info files.DirEntry) bool {
	return info.IsDir() && (info.Name() == ".git" || info.Name() == ".idea" || info.Name() == "node_modules")
}

// WalkFiles walks the file system rooted at root, calling f for each file. This is
// a less ceremonious version of filepath.WalkDir where only file paths (not dirs)
// are passed to the callback, and where directories that should commonly  be ignored
// (.git, node_modules, etc.) are skipped.
func WalkFiles(root string, f func(path string) error) error {
	return filepath.WalkDir(root, func(path string, info os.DirEntry, _ error) error { // nolint:wrapcheck
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
