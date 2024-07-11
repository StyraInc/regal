# METADATA
# description: '`with` used outside test context'
package regal.rules.performance["with-outside-test-context"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule_index, rule in input.rules
	some expr_index, expr in ast.exprs[rule_index]

	expr["with"]
	not strings.any_prefix_match(ast.ref_to_string(rule.head.ref), {"test_", "todo_test"})

	violation := result.fail(rego.metadata.chain(), result.location(expr["with"][0]))
}
