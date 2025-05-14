# METADATA
# description: Iteration in top-level assignment
package regal.rules.bugs["top-level-iteration"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some i
	input.rules[i].head.value.type == "ref"
	rule := input.rules[i]

	# skip if vars in the ref head
	count([part |
		some i, part in rule.head.ref
		i > 0
		part.type == "var"
	]) == 0

	some part in array.slice(rule.head.value.value, 1, 100)

	part.type == "var"

	_illegal_value_ref(part.value, rule, ast.identifiers)

	# this is expensive, but the preconditions should ensure that
	# very few rules evaluate this far
	not _var_in_body(rule, part.value)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

_var_in_body(rule, value) if {
	walk(rule.body, [_, node])
	node.type == "var"
	node.value == value
}

_illegal_value_ref(value, rule, identifiers) if {
	not value in identifiers
	not _is_arg_or_input(value, rule)
}

_is_arg_or_input(value, rule) if value in ast.function_arg_names(rule)
_is_arg_or_input(value, _) if value[0].value == "input"
_is_arg_or_input("input", _)
