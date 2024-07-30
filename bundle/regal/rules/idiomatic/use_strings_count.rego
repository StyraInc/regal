# METADATA
# description: Use `strings.count` where possible
package regal.rules.idiomatic["use-strings-count"]

import rego.v1

import data.regal.ast
import data.regal.capabilities
import data.regal.result

# METADATA
# description: Missing capability for built-in function `strings.count`
# custom:
#   severity: warning
notices contains result.notice(rego.metadata.chain()) if not capabilities.has_object_keys

# METADATA
# description: flag calls to `count` where the first argument is a call to `indexof_n`
report contains violation if {
	some rule in input.rules

	ref := ast.refs[_][_]

	ref[0].value[0].type == "var"
	ref[0].value[0].value == "count"

	ref[1].type == "call"
	ref[1].value[0].value[0].type == "var"
	ref[1].value[0].value[0].value == "indexof_n"

	loc1 := result.location(ref[0])
	loc2 := result.ranged_location_from_text(ref[1])

	violation := result.fail(rego.metadata.chain(), object.union(loc1, {"location": {"end": loc2.location.end}}))
}
