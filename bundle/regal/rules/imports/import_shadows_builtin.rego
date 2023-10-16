# METADATA
# description: Import shadows built-in namespace
package regal.rules.imports["import-shadows-builtin"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

builtin_namespaces contains namespace if {
	some name in ast.builtin_names

	namespace := split(name, ".")[0]
}

report contains violation if {
	some imp in input.imports

	imp.path.value[0].value in {"data", "input"}

	name := significant_name(imp)
	name in builtin_namespaces

	# AST quirk: while we'd ideally provide the location of the *path component*,
	# there is no location data provided for aliases. In order to be consistent,
	# we'll just provide the location of the import.
	violation := result.fail(rego.metadata.chain(), result.location(imp))
}

significant_name(imp) := imp.alias

significant_name(imp) := regal.last(imp.path.value).value if not imp.alias
