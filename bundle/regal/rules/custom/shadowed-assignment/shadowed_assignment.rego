# METADATA
# description: Shadowed assignment
# related_resources:
#   - description: documentation
#     ref: https://www.openpolicyagent.org/projects/regal/rules/custom/shadowed-assignment
package regal.rules.custom["shadowed-assignment"]

import data.regal.ast
import data.regal.result

report contains violation if {
	var := ast.found.vars[_].assign[_]
	var.value in ast.identifiers

	violation := result.fail(rego.metadata.chain(), result.location(var))
}
