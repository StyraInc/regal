# METADATA
# description: Entrypoint can't be marked internal
package regal.rules.bugs["internal-entrypoint"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in ast.rules
	some annotation in rule.annotations

	annotation.entrypoint == true

	some i, part in rule.head.ref

	_any_internal(i, part)

	violation := result.fail(rego.metadata.chain(), result.ranged_location_from_text(part))
}

_any_internal(0, part) if startswith(part.value, "_")

_any_internal(_, part) if {
	part.type == "string"
	startswith(part.value, "_")
}
