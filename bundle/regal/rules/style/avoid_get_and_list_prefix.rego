package regal.rules.style

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

# METADATA
# title: avoid-get-and-list-prefix
# description: Avoid get_ and list_ prefix for rules and functions
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/avoid-get-and-list-prefix
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	strings.any_prefix_match(rule.head.name, {"get_", "list_"})

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}
