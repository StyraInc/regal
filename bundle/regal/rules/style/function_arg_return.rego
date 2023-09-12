# METADATA
# description: Function argument used for return value
package regal.rules.style["function-arg-return"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.result

cfg := config.for_rule("style", "function-arg-return")

except_functions := array.concat(object.get(cfg, "except-functions", []), ["print"])

part_to_string(ref) := ref.value if ref.type == "string"

part_to_string(ref) := "$" if ref.type != "string"

report contains violation if {
	walk(input.rules, [path, value])

	regal.last(path) == "terms"

	value[0].type == "ref"
	value[0].value[0].type == "var"

	fn_name_parts := array.concat([value[0].value[0].value], [s |
		some i, part in value[0].value
		i > 0
		s := part_to_string(part)
	])

	fn_name := concat(".", fn_name_parts)

	not contains(fn_name, "$")
	not fn_name in except_functions
	fn_name in ast.all_function_names

	ast.function_ret_in_args(fn_name, value)

	violation := result.fail(rego.metadata.chain(), result.location(regal.last(value)))
}
