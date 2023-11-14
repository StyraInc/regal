# METADATA
# description: Import shadows another import
package regal.rules.imports["import-shadows-import"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.result

# regular import
_ident(imported) := regal.last(imported.path.value).value if not imported.alias

# aliased import
_ident(imported) := imported.alias

_identifiers := [_ident(imported) | some imported in input.imports]

report contains violation if {
	some i, identifier in _identifiers

	identifier in array.slice(_identifiers, 0, i)

	violation := result.fail(rego.metadata.chain(), result.location(input.imports[i].path.value[0]))
}
