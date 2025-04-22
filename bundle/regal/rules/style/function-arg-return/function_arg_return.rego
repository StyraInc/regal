# METADATA
# description: Function argument used for return value
package regal.rules.style["function-arg-return"]

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	excluded_functions := {name | some name in config.rules.style["function-arg-return"]["except-functions"]}
	included_functions := (ast.all_function_names - excluded_functions) - {"print"}

	some fn
	ast.function_calls[_][fn].name in included_functions

	count(fn.args) > count(ast.all_functions[fn.name].decl.args)

	violation := result.fail(rego.metadata.chain(), result.location(regal.last(fn.args)))
}
