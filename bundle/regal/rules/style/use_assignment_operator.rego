package regal.rules.style

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

# Some cases blocked by https://github.com/StyraInc/regal/issues/6 - e.g:
#
# allow = true { true }
#
# f(x) = 5

# METADATA
# title: use-assignment-operator
# description: Prefer := over = for assignment
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/use-assignment-operator
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	rule["default"] == true
	not rule.head.assign

	violation := result.fail(rego.metadata.rule(), result.location(rule))
}

# METADATA
# title: use-assignment-operator
# description: Prefer := over = for assignment
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/use-assignment-operator
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	rule.head.key
	rule.head.value
	not rule.head.assign

	violation := result.fail(rego.metadata.rule(), result.location(rule.head.ref[0]))
}
