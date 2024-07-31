package fixes

import (
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"
)

// NewDefaultFixes returns a list of default fixes that are applied by the fix command.
// When a new fix is added, it should be added to this list.
func NewDefaultFixes() []Fix {
	return []Fix{
		&Fmt{
			NameOverride: "use-rego-v1",
			OPAFmtOpts: format.Opts{
				RegoVersion: ast.RegoV0CompatV1,
			},
		},
		&UseAssignmentOperator{},
		&NoWhitespaceComment{},
	}
}

// Fix is the interface that must be implemented by all fixes.
type Fix interface {
	// Name returns the unique name for the fix, this should correlate with the
	// violation title & rule name that the fix is meant to address.
	Name() string
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
