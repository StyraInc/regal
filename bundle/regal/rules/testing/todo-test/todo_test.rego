# METADATA
# description: TODO test encountered
package regal.rules.testing["todo-test"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in ast.rules

	startswith(ast.ref_to_string(rule.head.ref), "todo_test_")

	violation := result.fail(rego.metadata.chain(), result.location(rule.head))
}
