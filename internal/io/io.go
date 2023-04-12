package io

import (
	"encoding/json"
	"fmt"
	"io"
	files "io/fs"
	"log"
	"os"
	"strings"

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

// MustLoadRegalBundlePath loads bundle from path, exit on failure.
func MustLoadRegalBundlePath(path string) bundle.Bundle {
	regalBundle, err := LoadRegalBundlePath(path)
	if err != nil {
		log.Fatal(err)
	}

	return regalBundle
}

// MustJSON parses JSON from reader, exit on failure.
func MustJSON(from any) []byte {
	bs, err := json.MarshalIndent(from, "", "   ")
	if err != nil {
		log.Fatal(err)
	}

	return bs
}

// JSONRoundTrip convert any value to JSON and back again.
func JSONRoundTrip(from any, to any) error {
	bs, err := json.Marshal(from)
	if err != nil {
		return fmt.Errorf("failed JSON marshalling %w", err)
	}

	return json.Unmarshal(bs, to) //nolint:wrapcheck
}

// MustYAMLToMap creates a map from reader, expecting YAML content, or fail.
func MustYAMLToMap(from io.Reader) (m map[string]any) {
	if err := yaml.NewDecoder(from).Decode(&m); err != nil {
		log.Fatal(err)
	}

	return m
}

// CloseFileIgnore closes file ignoring errors, mainly for deferred cleanup.
func CloseFileIgnore(file *os.File) {
	_ = file.Close()
}

func ExcludeTestFilter() filter.LoaderFilter {
	return func(abspath string, info files.FileInfo, depth int) bool {
		return strings.HasSuffix(info.Name(), "_test.rego")
	}
}
