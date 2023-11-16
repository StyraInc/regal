# METADATA
# description: Import declared after rule
package regal.rules.imports["import-after-rule"]

import rego.v1

import data.regal.result

report contains violation if {
	first_rule_row := input.rules[0].location.row

	some imp in input.imports

	imp.location.row > first_rule_row

	violation := result.fail(rego.metadata.chain(), result.location(imp))
}
