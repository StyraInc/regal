# METADATA
# description: '`with` used outside test context'
package regal.rules.performance["with-outside-test-context"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some i, rule in input.rules
	some expr in ast.found.expressions[ast.rule_index_strings[i]]

	expr.with
	not strings.any_prefix_match(ast.ref_to_string(rule.head.ref), {"test_", "todo_test"})

	violation := result.fail(rego.metadata.chain(), result.location(expr.with[0]))
}
