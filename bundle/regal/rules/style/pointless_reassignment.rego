# METADATA
# description: Pointless reassignment of variable
package regal.rules.style["pointless-reassignment"]

import rego.v1

import data.regal.ast
import data.regal.result

# pointless reassignment in rule head
report contains violation if {
	some rule in ast.rules

	ast.generated_body(rule)

	rule.head.value.type == "var"
	count(rule.head.ref) == 1

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}

# pointless reassignment in rule body
report contains violation if {
	some call in ast.all_refs

	call[0].value[0].type == "var"
	call[0].value[0].value == "assign"

	call[2].type == "var"

	violation := result.fail(rego.metadata.chain(), result.location(call))
}

assign_calls contains call if {
	some call in ast.all_refs

	call[0].value[0].type == "var"
	call[0].value[0].value == "assign"
}
