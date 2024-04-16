package fixes

import "github.com/open-policy-agent/opa/ast"

type Fix interface {
	Key() string
	Fix(in []byte, opts *RuntimeOptions) (bool, []byte, error)
}

type RuntimeOptions struct {
	Metadata  RuntimeMetadata
	Locations []ast.Location
}

type RuntimeMetadata struct {
	Filename string
}
