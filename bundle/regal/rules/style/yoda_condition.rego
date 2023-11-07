# METADATA
# description: Yoda condition
package regal.rules.style["yoda-condition"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	walk(input.rules, [_, value])

	value[0].type == "ref"
	value[0].value[0].type == "var"
	value[0].value[0].value in {"equal", "neq"} # perhaps add more operators here?
	value[1].type in ast.scalar_types
	not value[2].type in ast.scalar_types
	not ref_with_vars(value[2].value)

	violation := result.fail(rego.metadata.chain(), result.location(value))
}

ref_with_vars(ref) if {
	count(ref) > 2
	some i, part in ref
	i > 0
	part.type == "var"
}
