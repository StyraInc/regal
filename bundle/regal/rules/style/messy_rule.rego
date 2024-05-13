# METADATA
# description: Messy incremental rule
package regal.rules.style["messy-rule"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some i, rule1 in input.rules

	cur_name := ast.ref_to_string(rule1.head.ref)

	some j, rule2 in input.rules

	j > i

	nxt_name := ast.ref_to_string(rule2.head.ref)
	cur_name == nxt_name

	previous_name := ast.ref_to_string(input.rules[j - 1].head.ref)
	previous_name != nxt_name

	violation := result.fail(rego.metadata.chain(), result.location(rule2))
}
