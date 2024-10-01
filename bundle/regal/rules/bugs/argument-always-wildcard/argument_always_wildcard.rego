# METADATA
# description: Argument is always a wildcard
package regal.rules.bugs["argument-always-wildcard"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	some name, functions in _function_groups

	fn := _any_member(functions)

	some pos in numbers.range(0, count(fn.head.args) - 1)

	every function in functions {
		function.head.args[pos].type == "var"
		startswith(function.head.args[pos].value, "$")
	}

	not _function_name_excepted(config.for_rule("bugs", "argument-always-wildcard"), name)

	violation := result.fail(rego.metadata.chain(), result.location(fn.head.args[pos]))
}

_function_groups[name] contains fn if {
	some fn in ast.functions

	name := ast.ref_to_string(fn.head.ref)
}

_function_name_excepted(cfg, name) if regex.match(cfg["except-function-name-pattern"], name)

_any_member(s) := [x | some x in s][0]
