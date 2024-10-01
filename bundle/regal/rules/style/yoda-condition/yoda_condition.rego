# METADATA
# description: Yoda condition
package regal.rules.style["yoda-condition"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	value := ast.found.refs[_][_]

	value[0].value[0].type == "var"
	value[0].value[0].value in {"equal", "neq"} # perhaps add more operators here?
	value[1].type in ast.scalar_types

	not value[2].type in ast.scalar_types
	not _ref_with_vars(value[2].value)

	violation := result.fail(rego.metadata.chain(), result.infix_expr_location(value))
}

_ref_with_vars(ref) if {
	count(ref) > 2
	some i, part in ref
	i > 0
	part.type == "var"
}
