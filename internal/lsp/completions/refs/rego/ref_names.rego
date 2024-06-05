package lsp.completions

import rego.v1

import data.regal.ast

# ref_names returns a list of ref names that are used in the module.
# built-in functions are not included as they are provided by another completions provider.
# imports are not included as we need to use the imported_identifier instead
# (i.e. maybe an alias).
ref_names contains name if {
	some ref in ast.all_refs

	name := ast.ref_to_string(ref.value)

	not name in ast.builtin_functions_called
	not name in imports
}

# if a user has imported data.foo, then foo should be suggested.
# if they have imported data.foo as bar, then bar should be suggested.
# this also has the benefit of skipping future.* and rego.v1 as
# imported_identifiers will only match data.* and input.*
ref_names contains name if {
	some name in ast.imported_identifiers
}

# imports are not shown as we need to show the imported alias instead
imports contains ref if {
	some imp in ast.imports

	ref := ast.ref_to_string(imp.path.value)
}
