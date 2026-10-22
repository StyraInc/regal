# METADATA
# description: Redundant import of future keywords
# related_resources:
#   - description: documentation
#     ref: https://www.openpolicyagent.org/projects/regal/rules/imports/redundant-future-keywords
package regal.rules.imports["redundant-future-keywords"]

import data.regal.capabilities
import data.regal.config
import data.regal.result

default _future_keywords := []

_future_keywords := config.capabilities.future_keywords

_future_keywords := capabilities.provided.future_keywords if {
	not config.capabilities.future_keywords
}

report contains violation if {
	some imported in input.imports

	imported.path.type == "ref"

	count(imported.path.value) == 3

	imported.path.value[0].type == "var"
	imported.path.value[0].value == "future"
	imported.path.value[1].type == "string"
	imported.path.value[1].value == "keywords"

	keyword := imported.path.value[2].value

	not keyword in _future_keywords

	violation := result.fail(rego.metadata.chain(), result.location(imported.path.value[0]))
}
