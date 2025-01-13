# METADATA
# description: Missing entrypoint annotation
package regal.rules.idiomatic["no-defined-entrypoint"]

import data.regal.ast
import data.regal.result
import data.regal.util

# METADATA
# description: |
#   collects `entrypoint: true` annotations from any given module
aggregate contains entry if {
	some annotation in ast.annotations
	annotation.entrypoint == true

	entry := result.aggregate(rego.metadata.chain(), {"entrypoint": util.to_location_object(annotation.location)})
}

# METADATA
# schemas:
#   - input: schema.regal.aggregate
aggregate_report contains violation if {
	count(input.aggregate) == 0

	violation := result.fail(rego.metadata.chain(), {})
}
