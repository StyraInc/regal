# METADATA
# description: TODO test encountered
package regal.rules.testing["todo-test"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.result

report contains violation if {
	some rule in input.rules

	startswith(rule.head.name, "todo_test_")

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
