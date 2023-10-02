# METADATA
# description: Use `some` to declare output variables
package regal.rules.idiomatic["use-some-for-output-vars"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	# can't use ast.all_refs here as we need to
	# refer to the `rule` below
	some rule in input.rules

	walk(rule, [_, value])

	value.type == "ref"
	ref := value.value

	some i, elem in ref

	# first item can't be a loop or key ref
	i != 0
	elem.type == "var"
	not startswith(elem.value, "$")

	scope := ast.find_names_in_scope(rule, elem.location)
	not elem.value in scope

	violation := result.fail(rego.metadata.chain(), result.location(elem))
}
