// Copyright 2021 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

//go:build go1.16
// +build go1.16

package capabilities

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/open-policy-agent/opa/util"

	"github.com/open-policy-agent/opa/ast"
)

// FS contains the embedded capabilities/ directory of the built version,
// which has all the capabilities of previous versions:
// "v0.18.0.json" contains the capabilities JSON of version v0.18.0, etc
//
//go:embed *.json
var FS embed.FS

// LoadCapabilitiesJSON loads a JSON serialized capabilities structure from the reader r.
func LoadCapabilitiesJSON(r io.Reader) (*ast.Capabilities, error) {
	d := util.NewJSONDecoder(r)
	var c ast.Capabilities
	return &c, d.Decode(&c)
}

// LoadCapabilitiesVersion loads a JSON serialized capabilities structure from the specific version.
func LoadCapabilitiesVersion(version string) (*ast.Capabilities, error) {
	cvs, err := LoadCapabilitiesVersions()
	if err != nil {
		return nil, err
	}

	for _, cv := range cvs {
		if cv == version {
			cont, err := FS.ReadFile(cv + ".json")
			if err != nil {
				return nil, err
			}

			return LoadCapabilitiesJSON(bytes.NewReader(cont))
		}

	}
	return nil, fmt.Errorf("(Regal embedded EOPA capabilities library) no capabilities version found %v", version)
}

// LoadCapabilitiesFile loads a JSON serialized capabilities structure from a file.
func LoadCapabilitiesFile(file string) (*ast.Capabilities, error) {
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	return LoadCapabilitiesJSON(fd)
}

// LoadCapabilitiesVersions loads all capabilities versions
func LoadCapabilitiesVersions() ([]string, error) {
	ents, err := FS.ReadDir(".")
	if err != nil {
		return nil, err
	}

	capabilitiesVersions := make([]string, 0, len(ents))
	for _, ent := range ents {
		capabilitiesVersions = append(capabilitiesVersions, strings.Replace(ent.Name(), ".json", "", 1))
	}
	return capabilitiesVersions, nil
}
