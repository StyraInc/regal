package io

import (
	"encoding/json"
	files "io/fs"
	"log"
	"strings"

	"github.com/open-policy-agent/opa/bundle"
)

// LoadRegalBundle loads bundle embedded from policy and data directory
func LoadRegalBundle(fs files.FS) (bundle.Bundle, error) {
	embedLoader, err := bundle.NewFSLoader(fs)
	if err != nil {
		return bundle.Bundle{}, err
	}
	bundleLoader := embedLoader.WithFilter(func(abspath string, info files.FileInfo, depth int) bool {
		return strings.HasSuffix(info.Name(), "_test.rego")
	})

	return bundle.NewCustomReader(bundleLoader).
		WithSkipBundleVerification(true).
		WithProcessAnnotations(true).
		WithBundleName("regal").
		Read()
}

// MustLoadRegalBundle loads bundle embedded from policy and data directory, exit on failure
func MustLoadRegalBundle(fs files.FS) bundle.Bundle {
	regalBundle, err := LoadRegalBundle(fs)
	if err != nil {
		log.Fatal(err)
	}

	return regalBundle
}

// MustJSON marshal to JSON or exit
func MustJSON(x any) []byte {
	bytes, err := json.Marshal(x)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}
