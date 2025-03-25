# METADATA
# description: Repeated calls to `time.now_ns`
package regal.rules.bugs["time-now-ns-twice"]

import data.regal.ast
import data.regal.result

report contains violation if {
	# note: calls per _rule_index_, which is just what we want
	some calls in ast.function_calls

	time_now_calls := [call |
		some call in calls
		call.name == "time.now_ns"
	]

	count(time_now_calls) > 1

	some repeated in array.slice(time_now_calls, 1, count(time_now_calls))

	violation := result.fail(rego.metadata.chain(), result.location(repeated))
}
