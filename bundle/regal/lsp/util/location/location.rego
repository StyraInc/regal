# METADATA
# description: utility functions for dealing with location data in the LSP
package regal.lsp.util.location

import rego.v1

# METADATA
# description: turns an AST location (with `end`` attribute) into an LSP range
to_range(location) := {
	"start": {
		"line": location.row - 1,
		"character": location.col - 1,
	},
	"end": {
		"line": location.end.row - 1,
		"character": location.end.col - 1,
	},
}
