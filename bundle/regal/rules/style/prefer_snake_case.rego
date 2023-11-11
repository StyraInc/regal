# METADATA
# description: Prefer snake_case for names
package regal.rules.style["prefer-snake-case"]

import rego.v1

import data.regal.ast
import data.regal.result
import data.regal.util

report contains violation if {
	some rule in input.rules
	some ref in ast.named_refs(rule.head.ref)
	not util.is_snake_case(ref.value)

	violation := result.fail(rego.metadata.chain(), result.location(ref))
}

report contains violation if {
	some var in ast.find_vars(input.rules)
	not util.is_snake_case(var.value)

	violation := result.fail(rego.metadata.chain(), result.location(var))
}
