# METADATA
# description: Variable name shadows built-in
package regal.rules.bugs["var-shadows-builtin"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	var := ast.found.vars[_][_][_]

	var.value in ast.builtin_namespaces

	violation := result.fail(rego.metadata.chain(), result.location(var))
}
