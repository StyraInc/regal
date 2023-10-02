# METADATA
# description: Constant condition
package regal.rules.bugs["constant-condition"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

# NOTE: The constant condition checks currently don't do nesting!
# Additionally, there are several other conditions that could be considered
# constant, or if not, redundant... so this rule should be expanded in time

_operators := {"equal", "gt", "gte", "lt", "lte", "neq"}

_rules_with_bodies := [rule |
	some rule in input.rules
	not probably_no_body(rule)
]

# i.e. allow {..}, or allow := true, which expands to allow = true { true }
probably_no_body(rule) if {
	count(rule.body) == 1
	rule.body[0].terms.type == "boolean"
	rule.body[0].terms.value == true
}

report contains violation if {
	some rule in _rules_with_bodies
	some expr in rule.body

	# We could probably include arrays and objects too, as a single compound value
	# is not very useful, but it's not as clear cut as scalars, as you could have
	# something like {"a": foo(input.x) == "bar"} which is not a constant condition,
	# however meaningless it may be. Maybe consider for another rule?
	expr.terms.type in ast.scalar_types

	violation := result.fail(rego.metadata.chain(), result.location(expr))
}

report contains violation if {
	some rule in _rules_with_bodies
	some expr in rule.body

	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value in _operators

	expr.terms[1].type in ast.scalar_types
	expr.terms[2].type in ast.scalar_types

	violation := result.fail(rego.metadata.chain(), result.location(expr))
}
