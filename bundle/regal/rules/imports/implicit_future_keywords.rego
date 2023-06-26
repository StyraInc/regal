# METADATA
# description: Use explicit future keyword imports
package regal.rules.imports["implicit-future-keywords"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.result

report contains violation if {
	some imported in input.imports

	imported.path.type == "ref"

	count(imported.path.value) == 2

	imported.path.value[0].type == "var"
	imported.path.value[0].value == "future"
	imported.path.value[1].type == "string"
	imported.path.value[1].value == "keywords"

	violation := result.fail(rego.metadata.chain(), result.location(imported.path.value[0]))
}
