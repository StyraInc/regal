package regal.rules.imports

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

# regular import
_ident(imported) := regal.last(path).value if {
	not imported.alias
	path := imported.path.value
}

# aliased import
_ident(imported) := imported.alias

_identifiers := [_ident(imported) |
	some imported in input.imports
]

# METADATA
# title: import-shadows-import
# description: Import shadows another import
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/import-shadows-import
# custom:
#   category: imports
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some i, identifier in _identifiers

	identifier in array.slice(_identifiers, 0, i)

	violation := result.fail(rego.metadata.rule(), result.location(input.imports[i].path.value[0]))
}
