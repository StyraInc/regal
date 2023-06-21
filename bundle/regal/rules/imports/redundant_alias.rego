package regal.rules.imports

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

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
