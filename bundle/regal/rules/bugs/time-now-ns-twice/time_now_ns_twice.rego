# METADATA
# description: Repeated calls to `time.now_ns`
package regal.rules.bugs["time-now-ns-twice"]

import data.regal.ast
import data.regal.result

report contains violation if {
	some rule_index
	ast.function_calls[rule_index][_].name == "time.now_ns"

	time_now_calls := [call |
		some call in ast.function_calls[rule_index]
		call.name == "time.now_ns"
	]

	some i, repeated in time_now_calls
	i > 0

	violation := result.fail(rego.metadata.chain(), result.location(repeated))
}
