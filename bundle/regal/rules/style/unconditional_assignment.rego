# METADATA
# description: Unconditional assignment in rule body
package regal.rules.style["unconditional-assignment"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

# Single-value rules
report contains violation if {
	some rule in input.rules

	# There's going to be a few cases where more expressions
	# are in the body and it's still "unconditional", like e.g
	# a `print` call.. but let's keep it simple for now
	count(rule.body) == 1

	# Multi-value rules dealt with separately
	not rule.head.key

	# Remove this and consider proper handling of else once
	# https://github.com/open-policy-agent/opa/issues/5777
	# is resolved
	not rule["else"]

	# Var assignment in rule head
	rule.head.value.type == "var"

	# `with` statements can't be moved to the rule head
	not rule.body[0]["with"]

	assignment_expr(rule.body[0].terms)

	# Of var declared in rule head
	rule.body[0].terms[1].type == "var"
	rule.body[0].terms[1].value == rule.head.value.value

	violation := result.fail(rego.metadata.chain(), result.location(rule.body[0].terms[1]))
}

# Multi-value rules
# Comments added only where this differs from the above report
report contains violation if {
	some rule in ast.rules

	count(rule.body) == 1

	# Multi-value rule
	rule.head.key.type == "var"

	not rule.body[0]["with"]

	assignment_expr(rule.body[0].terms)

	rule.body[0].terms[1].type == "var"
	rule.body[0].terms[1].value == rule.head.key.value

	violation := result.fail(rego.metadata.chain(), result.location(rule.body[0].terms[1]))
}

# Assignment using either = or :=
assignment_expr(terms) if {
	terms[0].type == "ref"
	terms[0].value[0].type == "var"
	terms[0].value[0].value in {"eq", "assign"}
}
