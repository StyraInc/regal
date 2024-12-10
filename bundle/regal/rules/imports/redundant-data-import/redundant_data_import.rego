# METADATA
# description: Redundant import of data
package regal.rules.imports["redundant-data-import"]

import data.regal.result

report contains violation if {
	some imported in input.imports

	count(imported.path.value) == 1

	imported.path.value[0].value == "data"

	violation := result.fail(rego.metadata.chain(), result.location(imported.path.value[0]))
}
