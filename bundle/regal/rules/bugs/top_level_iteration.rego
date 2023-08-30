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

	illegal_value_ref(last.value)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}

_rule_names := {ast.name(rule) | some rule in input.rules}

# regal ignore:external-reference
illegal_value_ref(value) if not value in _rule_names
