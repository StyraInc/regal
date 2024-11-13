# METADATA
# description: |
#   Test to ensure a custom rule that aggregated no data is still reported
# related_resources:
#   - description: issue
#     ref: https://github.com/StyraInc/regal/issues/1259
package custom.regal.rules.testcase.empty_aggregates

import rego.v1

import data.regal.result

aggregate contains result.aggregate(rego.metadata.chain(), {}) if {
	input.nope
}

aggregate_report contains violation if {
	count(input.aggregate) == 0

	violation := result.fail(rego.metadata.chain(), {})
}
