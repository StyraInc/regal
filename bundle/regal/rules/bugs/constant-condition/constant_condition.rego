# METADATA
# description: Constant condition
package regal.rules.bugs["constant-condition"]

import data.regal.ast
import data.regal.result

# NOTE: The constant condition checks currently don't do nesting!
# Additionally, there are several other conditions that could be considered
# constant, or if not, redundant... so this rule should be expanded in time

# METADATA
# description: single scalar value, like a lone `true` inside a rule body
# scope: rule
report contains violation if {
	terms := input.rules[_].body[_].terms

	# We could probably include arrays and objects too, as a single compound value
	# is not very useful, but it's not as clear cut as scalars, as you could have
	# something like {"a": foo(input.x) == "bar"} which is not a constant condition,
	# however meaningless it may be. Maybe consider for another rule?
	terms.type in ast.scalar_types

	violation := result.fail(rego.metadata.chain(), result.location(terms))
}

# METADATA
# description: two scalar values with a "boolean operator" between, like 1 == 1, or 2 > 1
# scope: rule
report contains violation if {
	operators := {"equal", "gt", "gte", "lt", "lte", "neq"}

	expr := input.rules[_].body[_]

	expr.terms[0].value[0].type == "var"
	expr.terms[0].value[0].value in operators

	expr.terms[1].type in ast.scalar_types
	expr.terms[2].type in ast.scalar_types

	violation := result.fail(rego.metadata.chain(), result.location(expr))
}
