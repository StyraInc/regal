# METADATA
# description: Iteration in top-level assignment
package regal.rules.bugs["top-level-iteration"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

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

_rule_names := {ast.name(rule) | some rule in input.rules}

_arg_names(rule) := {arg.value | some arg in rule.head.args}

_path(loc) := concat(".", {l.value | some l in loc})

illegal_value_ref(value, rule) if {
	# regal ignore:external-reference
	not value in _rule_names
	not is_arg_or_input(value, rule)
}

is_arg_or_input(value, rule) if {
	value in _arg_names(rule)
}

is_arg_or_input(value, rule) if {
	startswith(_path(value), "input.")
}
