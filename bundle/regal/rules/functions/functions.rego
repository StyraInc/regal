package regal.rules.functions

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.result

# METADATA
# title: input-or-data-reference
# description: Reference to input, data or rule ref in function body
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/input-or-data-reference
# custom:
#   category: functions
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules
	rule.head.args

	named_args := {arg.value | some arg in rule.head.args; arg.type == "var"}

	some expr in rule.body

	is_array(expr.terms)

	some term in expr.terms

	term.type == "var"
	not term.value in named_args

	violation := result.fail(rego.metadata.rule(), result.location(term))
}

# METADATA
# title: input-or-data-reference
# description: Reference to input, data or rule ref in function body
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/input-or-data-reference
# custom:
#   category: functions
report contains violation if {
	config.for_rule(rego.metadata.rule()).enabled == true

	some rule in input.rules
	rule.head.args

	named_args := {arg.value | some arg in rule.head.args; arg.type == "var"}

	some expr in rule.body

	is_object(expr.terms)

	terms := expr.terms.value
	terms[0].type == "var"
	not terms[0].value in named_args

	violation := result.fail(rego.metadata.rule(), result.location(terms[0]))
}
