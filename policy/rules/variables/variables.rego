package regal.rules.variables

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal

# METADATA
# title: STY-VARIABLES-001
# description: Unconditional assignment in rule body
# related_resources:
# - https://docs.styra.com/regal/rules/sty-variables-001
violation contains msg if {
	some rule in input.rules

    # Single expression in rule body
    # There's going to be a few cases where more expressions
    # are in the body and it still "unconditional", like e.g
    # a `print` call.. but let's keep it simple for now
	count(rule.body) == 1

    # Var assignment in rule head
	rule.head.value.type == "var"
	rule_head_var := rule.head.value.value

	# Which is an assignment (= or :=)
    terms := rule.body[0].terms
	terms[0].type == "ref"
	terms[0].value[0].type == "var"
	terms[0].value[0].value in {"eq", "assign"}

    # Of var declared in rule head
	terms[1].type == "var"
    terms[1].value == rule_head_var

    msg := regal.fail(rego.metadata.rule(), {})
}