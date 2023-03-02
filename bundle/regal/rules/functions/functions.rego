package regal.rules.functions

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal

# METADATA
# title: input-or-data-reference
# description: Reference to input or data in function body
# related_resources:
# - description: documentation
#   ref: https://docs.styra.com/regal/rules/input-or-data-reference
# custom:
#   category: functions
report contains violation if {
	regal.rule_config(rego.metadata.rule()).enabled == true

	some rule in input.rules
	rule.head.args
	some expr in rule.body

	terms := expr.terms.value
	terms[0].type == "var"
	terms[0].value in {"input", "data"}

	violation := regal.fail(rego.metadata.rule(), {})
}
