# METADATA
# description: Avoid using deprecated built-in functions
package regal.rules.bugs["deprecated-builtin"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.result

report contains violation if {
	deprecated_builtins := {
		"any", "all", "re_match", "net.cidr_overlap", "set_diff", "cast_array",
		"cast_set", "cast_string", "cast_boolean", "cast_null", "cast_object",
	}

	# if none of the deprecated built-ins are in the
	# capabilities for the target, bail out early
	any_deprecated_builtin(object.keys(config.capabilities.builtins), deprecated_builtins)

	some ref in ast.all_refs

	call := ref[0]

	ast.ref_to_string(call.value) in deprecated_builtins

	violation := result.fail(rego.metadata.chain(), result.ranged_location_from_text(call))
}

any_deprecated_builtin(caps_builtins, deprecated_builtins) if {
	some builtin in caps_builtins
	builtin in deprecated_builtins
}
