# METADATA
# description: Import declared after rule
package regal.rules.imports["import-after-rule"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.result

report contains violation if {
	first_rule_row := input.rules[0].location.row

	some imp in input.imports

	imp.location.row > first_rule_row

	violation := result.fail(rego.metadata.chain(), result.location(imp))
}
