# METADATA
# description: Use `some` to declare output variables
package regal.rules.idiomatic["use-some-for-output-vars"]

import rego.v1

import data.regal.ast
import data.regal.result
import data.regal.util

report contains violation if {
	some rule_index, i
	elem := ast.found.refs[rule_index][_].value[i]

	# first item can't be a loop or key ref
	i != 0
	elem.type == "var"
	not startswith(elem.value, "$")

	rule := input.rules[to_number(rule_index)]

	not elem.value in ast.find_names_in_scope(rule, elem.location)

	path := _location_path(rule, elem.location)

	not var_in_comprehension_body(elem, rule, path)

	violation := result.fail(rego.metadata.chain(), result.ranged_location_from_text(elem))
}

_location_path(rule, location) := path if walk(rule, [path, location])

var_in_comprehension_body(var, rule, path) if {
	some v in _comprehension_body_vars(rule, path)
	v.type == var.type
	v.value == var.value
}

_comprehension_body_vars(rule, path) := [vars |
	some parent_path in array.reverse(util.all_paths(path))

	node := object.get(rule, parent_path, {})

	node.type in {"arraycomprehension", "objectcomprehension", "setcomprehension"}

	vars := ast.find_vars(node.value.body)
][0]
