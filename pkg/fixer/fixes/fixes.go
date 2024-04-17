package fixes

import "github.com/open-policy-agent/opa/ast"

// Fix is the interface that must be implemented by all fixes.
type Fix interface {
	// Key returns the unique key for the fix, this should correlate with the
	// violation that the fix is meant to address.
	Key() string
	Fix(fc *FixCandidate, opts *RuntimeOptions) (bool, []byte, error)
}

// RuntimeOptions are the options that are passed to the Fix method when the Fix is executed.
// Location based fixes will have the locations populated by the caller.
type RuntimeOptions struct {
	Locations []ast.Location
}

type FixCandidate struct {
	Filename string
	Contents []byte
}
