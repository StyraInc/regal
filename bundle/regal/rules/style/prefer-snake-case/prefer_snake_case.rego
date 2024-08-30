# METADATA
# description: Prefer snake_case for names
package regal.rules.style["prefer-snake-case"]

import rego.v1

import data.regal.ast
import data.regal.result
import data.regal.util

report contains violation if {
	some rule in input.rules
	some part in ast.named_refs(rule.head.ref)
	not util.is_snake_case(part.value)

	violation := result.fail(rego.metadata.chain(), result.ranged_location_from_text(part))
}

report contains violation if {
	var := ast.found.vars[_][_][_]
	not util.is_snake_case(var.value)

	violation := result.fail(rego.metadata.chain(), result.ranged_location_from_text(var))
}
