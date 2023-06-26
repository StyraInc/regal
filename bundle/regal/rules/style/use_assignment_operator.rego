# METADATA
# description: Prefer := over = for assignment
package regal.rules.style["use-assignment-operator"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	# foo = "bar"
	some rule in input.rules
	not rule["default"]
	not rule.head.assign
	not rule.head.key
	not ast.implicit_boolean_assignment(rule)

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}

report contains violation if {
	# default foo = "bar"
	some rule in input.rules
	rule["default"] == true
	not rule.head.assign

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}

report contains violation if {
	# foo["bar"] = "baz"
	some rule in input.rules
	rule.head.key
	rule.head.value
	not rule.head.assign

	violation := result.fail(rego.metadata.chain(), result.location(rule.head.ref[0]))
}

report contains violation if {
	# foo(bar) = "baz"
	some rule in input.rules
	rule.head.args
	not rule.head.assign
	not ast.implicit_boolean_assignment(rule)

	violation := result.fail(rego.metadata.chain(), result.location(rule.head.ref[0]))
}
