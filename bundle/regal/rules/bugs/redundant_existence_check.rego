# METADATA
# description: Redundant existence check
package regal.rules.bugs["redundant-existence-check"]

import rego.v1

import data.regal.ast
import data.regal.result

# METADATA
# description: check rule bodies for redundant existence checks
report contains violation if {
	some rule_index, rule in input.rules
	some expr_index, expr in ast.exprs[rule_index]

	expr.terms.type == "ref"

	not expr["with"]

	ast.static_ref(expr.terms)

	ref_str := ast.ref_to_string(expr.terms.value)
	next_expr := rule.body[expr_index + 1]

	some term in next_expr.terms

	ast.ref_to_string(term.value) == ref_str

	violation := result.fail(rego.metadata.chain(), result.ranged_location_from_text(expr))
}

# METADATA
# description: check for redundant existence checks in rule head assignment
report contains violation if {
	some rule_index, rule in input.rules

	rule.head.value.type == "ref"

	ref_str := ast.ref_to_string(rule.head.value.value)

	some expr in ast.exprs[rule_index]

	expr.terms.type == "ref"
	ast.ref_to_string(expr.terms.value) == ref_str

	violation := result.fail(rego.metadata.chain(), result.ranged_location_from_text(expr.terms))
}
