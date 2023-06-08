package regal.rules.idiomatic

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

# METADATA
# title: custom-has-key-construct
# description: Custom function may be replaced by `in` and `object.keys`
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/custom-has-key-construct
# custom:
#   category: idiomatic
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	rule.head.args

	arg_names := [arg.value | some arg in rule.head.args]

	count(rule.body) == 1

	terms := rule.body[0].terms

	terms[0].value[0].type == "var"
	terms[0].value[0].value == "eq"

	[var, ref] := normalize_eq_terms(terms)

	ref.value[0].type == "var"
	ref.value[0].value in arg_names
	ref.value[1].type == "var"
	ref.value[1].value in arg_names

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}

# METADATA
# description: Normalize var to always always be on the left hand side
normalize_eq_terms(terms) := [terms[1], terms[2]] if {
	terms[1].type == "var"
	terms[1].value == "$0"
	terms[2].type == "ref"
}

normalize_eq_terms(terms) := [terms[2], terms[1]] if {
	terms[1].type == "ref"
	terms[2].type == "var"
	terms[2].value == "$0"
}
