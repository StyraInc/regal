# METADATA
# description: Collect data in aggregates and validate it
package custom.regal.rules.testcase["aggregates"]

import rego.v1

import data.regal.result

aggregate contains result.aggregate(rego.metadata.chain(), {})

aggregate_report contains violation if {
	not two_files_processed

    violation := result.fail(rego.metadata.chain(), {})
}

two_files_processed if count(input.aggregate) == 2
