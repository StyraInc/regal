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

	violation := result.fail(rego.metadata.chain(), _location(rule, part))
}

report contains violation if {
	var := ast.found.vars[_][_][_]
	not util.is_snake_case(var.value)

	violation := result.fail(rego.metadata.chain(), result.ranged_location_from_text(var))
}

_location(_, part) := result.ranged_location_from_text(part) if part.location

# workaround until https://github.com/open-policy-agent/opa/issues/6860
# is fixed and we can trust that location is included for all ref parts
_location(rule, part) := result.ranged_location_from_text(rule.head) if not part.location
