# METADATA
# description: Use `some` to declare output variables
package regal.rules.idiomatic["use-some-for-output-vars"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule_index, i
	elem := ast.found.refs[rule_index][_].value[i]

	# first item can't be a loop or key ref
	i != 0
	elem.type == "var"
	not startswith(elem.value, "$")
	not elem.value in ast.find_names_in_scope(input.rules[to_number(rule_index)], elem.location)

	violation := result.fail(rego.metadata.chain(), result.location(elem))
}
