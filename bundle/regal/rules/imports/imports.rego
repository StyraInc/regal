package regal.rules.imports

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

_identifiers := [_ident(imported) |
	some imported in input.imports
]

# regular import
_ident(imported) := regal.last(path).value if {
	not imported.alias
	path := imported.path.value
}

# aliased import
_ident(imported) := imported.alias

# METADATA
# title: implicit-future-keywords
# description: Use explicit future keyword imports
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/implicit-future-keywords
# custom:
#   category: imports
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some imported in input.imports

	imported.path.type == "ref"

	count(imported.path.value) == 2

	imported.path.value[0].type == "var"
	imported.path.value[0].value == "future"
	imported.path.value[1].type == "string"
	imported.path.value[1].value == "keywords"

	violation := result.fail(rego.metadata.rule(), result.location(imported.path.value[0]))
}

# METADATA
# title: avoid-importing-input
# description: Avoid importing input
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/avoid-importing-input
# custom:
#   category: imports
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some imported in input.imports

	imported.path.value[0].value == "input"

	# Allow aliasing input, eg `import input as tfplan`:
	not _aliased_input(imported)

	violation := result.fail(rego.metadata.rule(), result.location(imported.path.value[0]))
}

_aliased_input(imported) if {
	count(imported.path.value) == 1
	imported.alias
}

# METADATA
# title: redundant-data-import
# description: Redundant import of data
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/redundant-data-import
# custom:
#   category: imports
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some imported in input.imports

	count(imported.path.value) == 1

	imported.path.value[0].value == "data"

	violation := result.fail(rego.metadata.rule(), result.location(imported.path.value[0]))
}

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

# METADATA
# title: redundant-alias
# description: Redundant alias
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/redundant-alias
# custom:
#   category: imports
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some imported in input.imports

	regal.last(imported.path.value).value == imported.alias

	violation := result.fail(rego.metadata.rule(), result.location(imported.path.value[0]))
}
