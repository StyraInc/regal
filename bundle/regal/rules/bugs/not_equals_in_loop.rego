# METADATA
# description: Use of != in loop
package regal.rules.bugs["not-equals-in-loop"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.result

report contains violation if {
	some rule in input.rules
	some expr in rule.body

	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value == "neq"

	some neq_term in array.slice(expr.terms, 1, count(expr.terms))
	neq_term.type == "ref"

	some i
	neq_term.value[i].type == "var"
	startswith(neq_term.value[i].value, "$")

	violation := result.fail(rego.metadata.chain(), result.location(expr.terms[0]))
}
