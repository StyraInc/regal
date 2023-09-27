# METADATA
# description: Non-boolean return value unused
package regal.rules.bugs["unused-return-value"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule in input.rules
	some expr in rule.body

	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"

	ref_name := expr.terms[0].value[0].value
	ref_name in ast.builtin_names

	data.regal.opa.builtins[ref_name].result.type != "boolean"

	# no violation if the return value is declared as the last function argument
	# see the function-arg-return rule for *that* violation
	not ast.function_ret_in_args(ref_name, expr.terms)

	violation := result.fail(rego.metadata.chain(), result.location(expr.terms[0]))
}
