# METADATA
# description: Argument is always a wildcard
package regal.rules.bugs["argument-always-wildcard"]

import rego.v1

import data.regal.ast
import data.regal.result

function_groups[name] contains fn if {
	some fn in ast.functions

	name := ast.ref_to_string(fn.head.ref)
}

report contains violation if {
	some functions in function_groups

	fn := any_member(functions)

	some pos in numbers.range(0, count(fn.head.args) - 1)

	every function in functions {
		function.head.args[pos].type == "var"
		startswith(function.head.args[pos].value, "$")
	}

	violation := result.fail(rego.metadata.chain(), result.location(fn.head.args[pos]))
}

any_member(s) := [x | some x in s][0]
