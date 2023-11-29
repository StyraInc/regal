# METADATA
# description: Redundant existence check
package regal.rules.bugs["redundant-existence-check"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules
	some i, expr in rule.body

	expr.terms.type == "ref"

	static_ref(expr.terms)

	ref_str := ast.ref_to_string(expr.terms.value)

	next_expr := rule.body[i + 1]

	some term in next_expr.terms

	ast.ref_to_string(term.value) == ref_str

	violation := result.fail(rego.metadata.chain(), result.location(expr))
}

static_ref(ref) if {
	every t in array.slice(ref.value, 1, count(ref.value)) {
		t.type == "string"
	}
}
