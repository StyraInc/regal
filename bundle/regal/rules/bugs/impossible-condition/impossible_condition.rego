# METADATA
# description: Conflicting conditions in rule body
# related_resources:
#   - description: documentation
#     ref: https://www.openpolicyagent.org/projects/regal/rules/bugs/impossible-condition
package regal.rules.bugs["impossible-condition"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules

	some i, cond1 in _equality_conditions(rule)
	some j, cond2 in _equality_conditions(rule)

	i < j

	# Both conditions compare the same target (variable or reference)
	cond1.target == cond2.target

	# But compare against different constant values
	cond1.value != cond2.value

	violation := result.fail(rego.metadata.chain(), result.location(cond2.location))
}

_equality_conditions(rule) := [cond |
	some expr in rule.body

	# Ignore expressions with `with` modifiers, as they alter mock data/scope
	not expr.with

	is_array(expr.terms)
	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value == "equal"

	cond := _extract_equality(expr.terms[1], expr.terms[2], expr.location)
]

_extract_equality(left, right, loc) := {
	"target": _target_repr(left),
	"value": right.value,
	"location": loc,
} if {
	_is_target(left)
	right.type in ast.scalar_types
}

_extract_equality(left, right, loc) := {
	"target": _target_repr(right),
	"value": left.value,
	"location": loc,
} if {
	_is_target(right)
	left.type in ast.scalar_types
}

_is_target(term) if term.type == "var"

_is_target(term) if {
	term.type == "ref"
	ast.static_ref(term)
}

_target_repr(term) := term.value if term.type == "var"

_target_repr(term) := ast.ref_to_string(term.value) if term.type == "ref"
