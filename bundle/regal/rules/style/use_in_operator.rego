# METADATA
# description: Use in to check for membership
package regal.rules.style["use-in-operator"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.result

report contains violation if {
	some expr in eq_exprs

	expr.terms[1].type in {"array", "boolean", "object", "null", "number", "set", "string", "var"}
	expr.terms[2].type == "ref"

	last := regal.last(expr.terms[2].value)

	last.type == "var"
	startswith(last.value, "$")

	violation := result.fail(rego.metadata.chain(), result.location(expr.terms[2].value[0]))
}

report contains violation if {
	some expr in eq_exprs

	expr.terms[1].type == "ref"
	expr.terms[2].type in {"array", "boolean", "object", "null", "number", "set", "string", "var"}

	last := regal.last(expr.terms[1].value)

	last.type == "var"
	startswith(last.value, "$")

	violation := result.fail(rego.metadata.chain(), result.location(expr.terms[1].value[0]))
}

eq_exprs contains expr if {
	some rule in input.rules
	some expr in rule.body

	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value == "equal"
}
