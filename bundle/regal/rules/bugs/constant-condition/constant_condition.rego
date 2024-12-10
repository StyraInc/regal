# METADATA
# description: Constant condition
package regal.rules.bugs["constant-condition"]

import data.regal.ast
import data.regal.result

# NOTE: The constant condition checks currently don't do nesting!
# Additionally, there are several other conditions that could be considered
# constant, or if not, redundant... so this rule should be expanded in time

_operators := {"equal", "gt", "gte", "lt", "lte", "neq"}

_rules_with_bodies := [rule |
	some rule in input.rules
	rule.body
]

# METADATA
# description: single scalar value, like a lone `true` inside a rule body
# scope: rule
report contains violation if {
	expr := _rules_with_bodies[_].body[_]

	# We could probably include arrays and objects too, as a single compound value
	# is not very useful, but it's not as clear cut as scalars, as you could have
	# something like {"a": foo(input.x) == "bar"} which is not a constant condition,
	# however meaningless it may be. Maybe consider for another rule?
	expr.terms.type in ast.scalar_types

	violation := result.fail(rego.metadata.chain(), result.location(expr.terms))
}

# METADATA
# description: two scalar values with a "boolean operator" between, like 1 == 1, or 2 > 1
# scope: rule
report contains violation if {
	expr := _rules_with_bodies[_].body[_]

	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value in _operators

	expr.terms[1].type in ast.scalar_types
	expr.terms[2].type in ast.scalar_types

	violation := result.fail(rego.metadata.chain(), result.location(expr))
}
