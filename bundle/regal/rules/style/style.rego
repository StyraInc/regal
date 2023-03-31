package regal.rules.style

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

# METADATA
# title: prefer-snake-case
# description: Prefer snake_case for names
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/prefer-snake-case
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules
	not util.is_snake_case(rule.head.name)

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}

# METADATA
# title: prefer-snake-case
# description: Prefer snake_case for names
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/prefer-snake-case
# custom:
#   category: style
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some var in ast.find_vars(input.rules)
	not util.is_snake_case(var.value)

	violation := result.fail(rego.metadata.rule(), result.location(var))
}
