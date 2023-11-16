# METADATA
# description: Use of != in loop
package regal.rules.bugs["not-equals-in-loop"]

import rego.v1

import data.regal.result

report contains violation if {
	expr := input.rules[_].body[_]

	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value == "neq"

	some neq_term in array.slice(expr.terms, 1, count(expr.terms))
	neq_term.type == "ref"

	some value in neq_term.value
	value.type == "var"
	startswith(value.value, "$")

	violation := result.fail(rego.metadata.chain(), result.location(expr.terms[0]))
}
