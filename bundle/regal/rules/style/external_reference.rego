# METADATA
# description: Reference to input, data or rule ref in function body
package regal.rules.style["external-reference"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some fn in ast.functions

	named_args := {arg.value | some arg in fn.head.args; arg.type == "var"}
	own_vars := {v.value | some v in ast.find_vars(fn.body)}

	allowed_refs := named_args | own_vars

	some expr in fn.body
	some term in expr_terms(expr.terms)

	term.type == "var"
	not term.value in allowed_refs
	not startswith(term.value, "$")

	violation := result.fail(rego.metadata.chain(), result.location(term))
}

expr_terms(terms) := terms if is_array(terms)

expr_terms(terms) := [terms.value[0]] if is_object(terms)
