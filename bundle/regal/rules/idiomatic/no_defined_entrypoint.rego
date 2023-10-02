# METADATA
# description: Missing entrypoint annotation
package regal.rules.idiomatic["no-defined-entrypoint"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.result

aggregate contains entry if {
	some annotation in input.annotations
	annotation.entrypoint == true

	entry := result.aggregate(rego.metadata.chain(), {"entrypoint": annotation.location})
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
aggregate_report contains violation if {
	count(input.aggregate) == 0

	violation := result.fail(rego.metadata.chain(), {})
}
