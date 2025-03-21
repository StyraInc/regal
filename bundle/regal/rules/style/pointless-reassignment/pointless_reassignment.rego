# METADATA
# description: Pointless reassignment of variable
package regal.rules.style["pointless-reassignment"]

import data.regal.ast
import data.regal.result

# pointless reassignment in rule head
report contains violation if {
	some rule in ast.rules

	not rule.body

	rule.head.value.type == "var"
	count(rule.head.ref) == 1

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}

# pointless reassignment in rule body
report contains violation if {
	some rule in input.rules
	some expr in rule.body

	not expr["with"]

	[lhs, rhs] := ast.assignment_terms(expr)

	lhs.type == "var"
	rhs.type == "var"

	violation := result.fail(rego.metadata.chain(), result.infix_expr_location(expr.terms))
}
