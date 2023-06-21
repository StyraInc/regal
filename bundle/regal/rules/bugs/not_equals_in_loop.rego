package regal.rules.bugs

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

# METADATA
# title: not-equals-in-loop
# description: Use of != in loop
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/not-equals-in-loop
# custom:
#   category: bugs
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

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

	violation := result.fail(rego.metadata.rule(), result.location(expr.terms[0]))
}
