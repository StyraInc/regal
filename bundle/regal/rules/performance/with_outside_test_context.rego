# METADATA
# description: '`with` used outside test context'
package regal.rules.performance["with-outside-test-context"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules
	some expr in rule.body

	expr["with"]
	not strings.any_prefix_match(ast.name(rule), {"test_", "todo_test"})

	violation := result.fail(rego.metadata.chain(), result.location(expr["with"][0]))
}
