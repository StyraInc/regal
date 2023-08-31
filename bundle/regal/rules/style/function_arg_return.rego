# METADATA
# description: Function argument used for return value
package regal.rules.style["function-arg-return"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	cfg := config.for_rule("style", "function-arg-return")

	except_functions := array.concat(
		object.get(cfg, "except-functions", []),
		["print"],
	)

	# rule ignoring itself :)
	# regal ignore:function-arg-return,unused-return-value
	walk(input.rules, [path, value])

	regal.last(path) == "terms"

	value[0].type == "ref"
	value[0].value[0].type == "var"

	fn_name := value[0].value[0].value

	not fn_name in except_functions
	fn_name in ast.all_function_names

	ast.function_ret_in_args(fn_name, value)

	violation := result.fail(rego.metadata.chain(), result.location(regal.last(value)))
}
