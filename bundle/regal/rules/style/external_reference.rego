# METADATA
# description: Reference to input, data or rule ref in function body
package regal.rules.style["external-reference"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules
	rule.head.args

	named_args := {arg.value | some arg in rule.head.args; arg.type == "var"}
	own_vars := {v.value | some v in ast.find_vars(rule.body)}

	allowed_refs := named_args | own_vars

	some expr in rule.body

	is_array(expr.terms)

	some term in expr.terms

	term.type == "var"
	not term.value in allowed_refs

	violation := result.fail(rego.metadata.chain(), result.location(term))
}

report contains violation if {
	some rule in input.rules
	rule.head.args

	named_args := {arg.value | some arg in rule.head.args; arg.type == "var"}
	own_vars := {v.value | some v in ast.find_vars(rule.body)}

	allowed_refs := named_args | own_vars

	some expr in rule.body

	is_object(expr.terms)

	terms := expr.terms.value
	terms[0].type == "var"
	not terms[0].value in allowed_refs

	violation := result.fail(rego.metadata.chain(), result.location(terms[0]))
}
