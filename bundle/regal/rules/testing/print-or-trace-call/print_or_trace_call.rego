# METADATA
# description: Call to print or trace function
package regal.rules.testing["print-or-trace-call"]

import rego.v1

import data.regal.ast
import data.regal.result
import data.regal.util

report contains violation if {
	# skip iteration of refs if no print or trace calls are registered
	util.intersects(ast.builtin_functions_called, {"print", "trace"})

	ref := ast.found.refs[_][_]

	ref[0].value[0].type == "var"
	ref[0].value[0].value in {"print", "trace"}

	violation := result.fail(rego.metadata.chain(), result.location(ref[0]))
}
