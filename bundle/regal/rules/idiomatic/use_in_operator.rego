# METADATA
# description: Use in to check for membership
package regal.rules.idiomatic["use-in-operator"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.result

report contains violation if {
	some terms in eq_exprs_terms

	terms[1].type in {"array", "boolean", "object", "null", "number", "set", "string", "var"}
	terms[2].type == "ref"

	last := regal.last(terms[2].value)

	last.type == "var"
	startswith(last.value, "$")

	violation := result.fail(rego.metadata.chain(), result.location(terms[2].value[0]))
}

report contains violation if {
	some terms in eq_exprs_terms

	terms[1].type == "ref"
	terms[2].type in {"array", "boolean", "object", "null", "number", "set", "string", "var"}

	last := regal.last(terms[1].value)

	last.type == "var"
	startswith(last.value, "$")

	violation := result.fail(rego.metadata.chain(), result.location(terms[1].value[0]))
}

eq_exprs_terms contains terms if {
	terms := input.rules[_].body[_].terms

	terms[0].type == "ref"
	terms[0].value[0].type == "var"
	terms[0].value[0].value == "equal"
}
