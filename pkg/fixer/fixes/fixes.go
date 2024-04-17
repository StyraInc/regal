package fixes

import "github.com/open-policy-agent/opa/ast"

// Fix is the interface that must be implemented by all fixes.
type Fix interface {
	// Key returns the unique key for the fix, this should correlate with the
	// violation that the fix is meant to address.
	Key() string
	Fix(fc *FixCandidate, opts *RuntimeOptions) ([]FixResult, error)
}

// RuntimeOptions are the options that are passed to the Fix method when the Fix is executed.
// Location based fixes will have the locations populated by the caller.
type RuntimeOptions struct {
	Locations []ast.Location
}

// FixCandidate is the input to a Fix method and represents a file in need of fixing.
type FixCandidate struct {
	Filename string
	Contents []byte
}

// FixResult is returned from the Fix method and contains the new contents or fix recommendations.
// In future this might support diff based updates, or renames.
type FixResult struct {
	Contents []byte
}
