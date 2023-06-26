# METADATA
# description: Call to print or trace function
package regal.rules.testing["print-or-trace-call"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.result

report contains violation if {
	some call in ast.find_builtin_calls(input)

	name := call[0].value[0].value
	name in {"print", "trace"}

	violation := result.fail(rego.metadata.chain(), result.location(call[0].value[0]))
}
