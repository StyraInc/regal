# METADATA
# description: Non-boolean return value unassigned
package regal.rules.bugs["unassigned-return-value"]

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	some expr in ast.exprs[_]

	expr.terms[0].type == "ref"
	expr.terms[0].value[0].type == "var"

	ref_name := expr.terms[0].value[0].value
	ref_name in ast.builtin_names

	# special case as the "result" of print is ""
	ref_name != "print"

	config.capabilities.builtins[ref_name].decl.result != "boolean"

	# no violation if the return value is declared as the last function argument
	# see the function-arg-return rule for *that* violation
	not ast.function_ret_in_args(ref_name, expr.terms)

	violation := result.fail(rego.metadata.chain(), result.location(expr.terms[0]))
}
