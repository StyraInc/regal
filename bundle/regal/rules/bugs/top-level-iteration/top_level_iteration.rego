# METADATA
# description: Iteration in top-level assignment
package regal.rules.bugs["top-level-iteration"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules

	# skip if vars in the ref head
	count([part |
		some i, part in rule.head.ref
		i > 0
		part.type == "var"
	]) == 0

	rule.head.value.type == "ref"

	last := regal.last(rule.head.value.value)

	last.type == "var"
	_illegal_value_ref(last.value, rule, ast.identifiers)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

_path(loc) := concat(".", {l.value | some l in loc})

_illegal_value_ref(value, rule, identifiers) if {
	not value in identifiers
	not _is_arg_or_input(value, rule)
}

_is_arg_or_input(value, rule) if value in ast.function_arg_names(rule)

_is_arg_or_input(value, _) if startswith(_path(value), "input.")

_is_arg_or_input("input", _)
