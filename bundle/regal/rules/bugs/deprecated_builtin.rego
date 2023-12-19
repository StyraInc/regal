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

	ref[0].value[0].type == "var"
	not ref[0].value[0].value in {"input", "data"}

	name := concat(".", [value |
		some part in ref[0].value
		value := part.value
	])

	name in deprecated_builtins

	violation := result.fail(rego.metadata.chain(), result.location(ref))
}

any_deprecated_builtin(caps_builtins, deprecated_builtins) if {
	some builtin in caps_builtins
	builtin in deprecated_builtins
}
