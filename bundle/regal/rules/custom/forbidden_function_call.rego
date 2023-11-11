# METADATA
# description: Forbidden function call
package regal.rules.custom["forbidden-function-call"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result

cfg := config.for_rule("custom", "forbidden-function-call")

any_forbidden_function_called if {
	some function in cfg["forbidden-functions"]
	function in ast.builtin_functions_called
}

report contains violation if {
	# avoid traversal if no forbidden function is called
	any_forbidden_function_called

	some ref in ast.all_refs

	name := ast.ref_to_string(ref[0].value)
	name in cfg["forbidden-functions"]

	violation := result.fail(rego.metadata.chain(), result.location(ref))
}
