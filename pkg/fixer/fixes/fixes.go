package fixes

import "github.com/open-policy-agent/opa/ast"

// Fix is the interface that must be implemented by all fixes.
type Fix interface {
	// Key returns the unique key for the fix, this should correlate with the
	// violation that the fix is meant to address.
	Key() string
	// WholeFile returns true if the fix operates on the whole file,
	// false if it operates on specific locations.
	WholeFile() bool
	Fix(in []byte, opts *RuntimeOptions) (bool, []byte, error)
}

// RuntimeOptions are the options that are passed to the Fix method when the Fix is executed.
// Location based fixes will have the locations populated by the caller.
type RuntimeOptions struct {
	Metadata  RuntimeMetadata
	Locations []ast.Location
}

type RuntimeMetadata struct {
	// Filename will be set on fixes that need the filename, this is sometimes needed
	// for error messages.
	Filename string
}
