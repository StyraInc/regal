# METADATA
# description: Messy incremental rule
package regal.rules.style["messy-rule"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some i, rule1 in input.rules

	cur_name := ast.ref_static_to_string(rule1.head.ref)

	# tests aren't really incremental rules, and other rules
	# will flag multiple rules with the same name
	not startswith(cur_name, "test_")
	not startswith(cur_name, "todo_test_")

	some j, rule2 in input.rules

	j > i

	nxt_name := ast.ref_static_to_string(rule2.head.ref)
	cur_name == nxt_name

	previous_name := ast.ref_static_to_string(input.rules[j - 1].head.ref)
	previous_name != nxt_name

	violation := result.fail(rego.metadata.chain(), result.location(rule2))
}
