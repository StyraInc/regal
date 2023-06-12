package regal.rules.imports

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

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
