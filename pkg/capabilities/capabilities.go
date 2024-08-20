// package capabilities provides convenient access to OPA capabilities
// definitions that are embedded within Regal.
package capabilities

import (
	"fmt"

	"github.com/open-policy-agent/opa/ast"

	eopa_caps "github.com/styrainc/enterprise-opa/capabilities"
)

const (
	engineOPA  = "opa"
	engineEOPA = "eopa"
)

// Lookup attempts to access the capabilities requested engine and version. It
// returns nil if the capability is not found in the embedded capabilities
// database.
func Lookup(engine, version string) (*ast.Capabilities, error) {
	switch engine {
	case engineOPA:
		return ast.LoadCapabilitiesVersion(version)
	case engineEOPA:
		return eopa_caps.LoadCapabilitiesVersion(version)
	default:
		return nil, fmt.Errorf("engine '%s' not present in embedded capabilities database", engine)
	}
}

// List returns a map with keys being Rego engine types, and values being lists
// of capabilities versions present in the embedded capabilities database for
// that version.
func List() (map[string][]string, error) {
	opaCaps, err := ast.LoadCapabilitiesVersions()
	if err != nil {
		return nil, err
	}

	eopaCaps, err := eopa_caps.LoadCapabilitiesVersions()
	if err != nil {
		return nil, err
	}

	return map[string][]string{
		engineOPA:  opaCaps,
		engineEOPA: eopaCaps,
	}, nil
}
