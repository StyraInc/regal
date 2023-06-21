# METADATA
# description: Prefer := over = for assignment
package regal.rules.style["use-assignment-operator"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.result

# Some cases blocked by https://github.com/StyraInc/regal/issues/6 - e.g:
#
# allow = true { true }
#
# f(x) = 5

todo_report contains violation if {
	# foo = "bar"
	some rule in input.rules
	not rule["default"]
	not rule.head.assign
	not possibly_implicit_assign(rule)

	# print(text_for_location(rule.head.location))

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}

empty_body(rule) if {
	rule.body == [{"index": 0, "terms": {"type": "boolean", "value": true}}]
}

possibly_implicit_assign(rule) if {
	rule.head.value == {"type": "boolean", "value": true}
}

report contains violation if {
	# default foo = "bar"
	some rule in input.rules
	rule["default"] == true
	not rule.head.assign

	violation := result.fail(rego.metadata.chain(), result.location(rule))
}

report contains violation if {
	# foo["bar"] = "baz"
	some rule in input.rules
	rule.head.key
	rule.head.value
	not rule.head.assign

	violation := result.fail(rego.metadata.chain(), result.location(rule.head.ref[0]))
}

text_for_location(location) := {"location": object.union(
	location,
	{"text": input.regal.file.lines[location.row - 1]},
)} if {
	location.row
} else := {"location": location}
