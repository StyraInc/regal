# METADATA
# description: Outside reference to internal rule or function
package regal.rules.bugs["leaked-internal-reference"]

import rego.v1

import data.regal.ast
import data.regal.result

report contains violation if {
	ref := ast.found.refs[_][_]

	contains(ast.ref_to_string(ref.value), "._")

	violation := result.fail(rego.metadata.chain(), result.location(ref))
}

report contains violation if {
	some imported in input.imports

	contains(ast.ref_to_string(imported.path.value), "._")

	violation := result.fail(rego.metadata.chain(), result.location(imported.path.value))
}
