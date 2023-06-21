package regal.rules.imports

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

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
