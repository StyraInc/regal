package regal.rules.functions

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal

# METADATA
# title: STY-FUNCTIONS-001
# description: Reference to input or data in function body
# related_resources:
# - https://docs.styra.com/regal/rules/sty-functions-001
violation contains msg if {
	some rule in input.rules
	rule.head.args
    some expr in rule.body

    terms := expr.terms.value
    terms[0].type == "var"
    terms[0].value in {"input", "data"}

	msg := regal.fail(rego.metadata.rule(), {})
}
