# METADATA
# description: Prefer snake_case for names
package regal.rules.style["prefer-snake-case"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result
import data.regal.util

report contains violation if {
	some rule in input.rules
	not util.is_snake_case(rule.head.name)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

report contains violation if {
	some var in ast.find_vars(input.rules)
	not util.is_snake_case(var.value)

	violation := result.fail(rego.metadata.chain(), result.location(var))
}
