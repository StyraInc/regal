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

// MustLoadRegalBundlePath loads bundle from path, exit on failure.
func MustLoadRegalBundlePath(path string) bundle.Bundle {
	regalBundle, err := LoadRegalBundlePath(path)
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
