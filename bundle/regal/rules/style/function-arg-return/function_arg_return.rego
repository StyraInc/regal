# METADATA
# description: Function argument used for return value
package regal.rules.style["function-arg-return"]

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	cfg := config.for_rule("style", "function-arg-return")

	excluded_functions := {name | some name in cfg["except-functions"]} | {"print"}
	included_functions := ast.all_function_names - excluded_functions

	some rule_index in ast.rule_index_strings
	some fn in ast.function_calls[rule_index]

	fn.name in included_functions

	count(fn.args) > count(ast.all_functions[fn.name].decl.args)

	violation := result.fail(rego.metadata.chain(), result.location(regal.last(fn.args)))
}
