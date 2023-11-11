# METADATA
# description: Iteration in top-level assignment
package regal.rules.bugs["top-level-iteration"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules

	rule.head.value.type == "ref"

	last := regal.last(rule.head.value.value)

	last.type == "var"
	illegal_value_ref(last.value, rule)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

_path(loc) := concat(".", {l.value | some l in loc})

illegal_value_ref(value, rule) if {
	# regal ignore:external-reference
	not value in ast.rule_and_function_names
	not is_arg_or_input(value, rule)
}

is_arg_or_input(value, rule) if value in ast.function_arg_names(rule)

is_arg_or_input(value, _) if startswith(_path(value), "input.")

# ideally would be able to just say
# is_arg_or_input("input", _)
# but the formatter rewrites that to
# is_arg_or_input("input", _) = true
# ...meh
is_arg_or_input("input", _) := true
