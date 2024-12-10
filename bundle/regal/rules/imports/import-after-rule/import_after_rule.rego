# METADATA
# description: Import declared after rule
package regal.rules.imports["import-after-rule"]

import data.regal.result
import data.regal.util

report contains violation if {
	first_rule_row := util.to_location_object(input.rules[0].location).row

	some imp in input.imports

	util.to_location_object(imp.location).row > first_rule_row

	violation := result.fail(rego.metadata.chain(), result.location(imp))
}
