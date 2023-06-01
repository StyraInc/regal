package regal.rules.bugs

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.opa
import data.regal.result

# We could probably include arrays and objects too, as a single compound value
# is not very useful, but it's not as clear cut as scalars, as you could have
# something like {"a": foo(input.x) == "bar"} which is not a constant condition,
# however meaningless it may be. Maybe consider for another rule?
_scalars := {"boolean", "null", "number", "string"}

_operators := {"equal", "gt", "gte", "lt", "lte", "neq"}

_rule_names := {name | name := input.rules[_].head.name}

_rules_with_bodies := [rule |
	some rule in input.rules
	not probably_no_body(rule)
]

# NOTE: The constant condition checks currently don't do nesting!
# Additionally, there are several other conditions that could be considered
# constant, or if not, redundant... so this rule should be expanded in time

# METADATA
# title: constant-condition
# description: Constant condition
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/constant-condition
# custom:
#   category: bugs
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in _rules_with_bodies
	some expr in rule.body

	expr.terms.type in _scalars

	violation := result.fail(rego.metadata.rule(), result.location(expr))
}

# METADATA
# title: constant-condition
# description: Constant condition
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/constant-condition
# custom:
#   category: bugs
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in _rules_with_bodies
	some expr in rule.body

	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value in _operators

	expr.terms[1].type in _scalars
	expr.terms[2].type in _scalars

	violation := result.fail(rego.metadata.rule(), result.location(expr))
}

# METADATA
# title: top-level-iteration
# description: Iteration in top-level assignment
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/top-level-iteration
# custom:
#   category: bugs
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules

	rule.head.value.type == "ref"

	last := regal.last(rule.head.value.value)
	last.type == "var"

	illegal_value_ref(last.value)

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}

_builtin_names := object.keys(opa.builtins)

# METADATA
# title: unused-return-value
# description: Non-boolean return value unused
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/unused-return-value
# custom:
#   category: bugs
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	some expr in rule.body

	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"

	ref_name := expr.terms[0].value[0].value
	ref_name in _builtin_names

	opa.builtins[ref_name].result.type != "boolean"

	violation := result.fail(rego.metadata.rule(), result.location(expr.terms[0]))
}

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

# regal ignore:external-reference
illegal_value_ref(value) if not value in _rule_names

# i.e. allow {..}, or allow := true, which expands to allow = true { true }
probably_no_body(rule) if {
	count(rule.body) == 1
	rule.body[0].terms.type == "boolean"
	rule.body[0].terms.value == true
}

# METADATA
# title: rule-shadows-builtin
# description: Rule name shadows built-in
# related_resources:
# - description: documentation
#   ref: $baseUrl/$category/rule-shadows-builtin
# custom:
#   category: bugs
report contains violation if {
	config.for_rule(rego.metadata.rule()).level != "ignore"

	some rule in input.rules
	rule.head.name in _builtin_names

	violation := result.fail(rego.metadata.rule(), result.location(rule.head))
}
