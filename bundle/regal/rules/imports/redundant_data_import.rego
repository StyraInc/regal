package regal.rules.imports

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

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
