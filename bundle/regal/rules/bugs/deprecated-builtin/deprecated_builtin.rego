# METADATA
# description: Avoid using deprecated built-in functions
package regal.rules.bugs["deprecated-builtin"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result
import data.regal.util

report contains violation if {
	deprecated_builtins := {
		"any", "all", "re_match", "net.cidr_overlap", "set_diff", "cast_array",
		"cast_set", "cast_string", "cast_boolean", "cast_null", "cast_object",
	}

	# bail out early if no the deprecated built-ins are in capabilities
	util.intersects(object.keys(config.capabilities.builtins), deprecated_builtins)

	ref := ast.found.refs[_][_]
	call := ref[0]

	ast.ref_to_string(call.value) in deprecated_builtins

	violation := result.fail(rego.metadata.chain(), result.location(call))
}
