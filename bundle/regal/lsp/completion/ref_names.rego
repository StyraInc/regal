package regal.lsp.completion

import data.regal.ast

# METADATA
# description: |
#   returns a list of ref names that are used in the module
#   built-in functions are not included as they are provided by another completions provider
# scope: document
ref_names contains name if {
	name := ast.ref_static_to_string(ast.found.refs[_][_].value)

	not name in ast.builtin_names
}

# if a user has imported data.foo, then foo should be suggested.
# if they have imported data.foo as bar, then bar should be suggested.
# this also has the benefit of skipping future.* and rego.v1 as
# imported_identifiers will only match data.* and input.*
ref_names contains name if {
	some name in ast.imported_identifiers
}
