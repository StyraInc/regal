# METADATA
# description: Call to print or trace function
package regal.rules.testing["print-or-trace-call"]

import rego.v1

import data.regal.ast
import data.regal.result

print_or_trace_called if {
	some name in {"print", "trace"}
	name in ast.builtin_functions_called
}

report contains violation if {
	# skip iteration of refs if no print or trace calls are registered
	print_or_trace_called

	some ref in ast.all_refs

	ref[0].value[0].type == "var"
	ref[0].value[0].value in {"print", "trace"}

	violation := result.fail(rego.metadata.chain(), result.location(ref[0]))
}
