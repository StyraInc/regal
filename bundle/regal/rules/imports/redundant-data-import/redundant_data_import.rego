# METADATA
# description: Redundant import of data
package regal.rules.imports["redundant-data-import"]

import data.regal.result

report contains violation if {
	path := input.imports[_].path.value

	count(path) == 1
	path[0].value == "data"

	violation := result.fail(rego.metadata.chain(), result.location(path[0]))
}
