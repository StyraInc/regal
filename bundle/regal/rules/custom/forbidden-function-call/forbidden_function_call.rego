# METADATA
# description: Forbidden function call
package regal.rules.custom["forbidden-function-call"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

report contains violation if {
	cfg := config.for_rule("custom", "forbidden-function-call")

	# avoid traversal if no forbidden function is called
	util.intersects(util.to_set(cfg["forbidden-functions"]), ast.builtin_functions_called)

	ref := ast.found.refs[_][_]
	name := ast.ref_to_string(ref[0].value)
	name in cfg["forbidden-functions"]

	violation := result.fail(rego.metadata.chain(), result.location(ref))
}
