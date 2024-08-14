# METADATA
# description: Duplicate rule
package regal.rules.bugs["duplicate-rule"]

import rego.v1

import data.regal.result
import data.regal.util

report contains violation if {
	some indices in duplicates

	first := indices[0]
	rest := array.slice(indices, 1, count(indices))

	dup_locations := [location |
		some index in rest
		location := util.to_location_object(input.rules[index].location)
	]

	violation := result.fail(rego.metadata.chain(), object.union(
		result.ranged_location_from_text(input.rules[first]),
		{"description": message(dup_locations)},
	))
}

message(locations) := sprintf("Duplicate rule found at line %d", [locations[0].row]) if count(locations) == 1

message(locations) := sprintf(
	"Duplicate rules found at lines %s",
	[concat(", ", [line |
		some location in locations
		line := sprintf("%d", [location.row])
	])],
) if {
	count(locations) > 1
}

rules_as_text := [base64.decode(util.to_location_object(rule.location).text) | some rule in input.rules]

duplicates contains indices if {
	# Remove whitespace from textual representation of rule and create a hash from the result.
	# This provides a decent, and importantly *cheap*, approximation of duplicates. We can then
	# parse the text of these suspected duplicate rules to get a more exact result.
	rules_hashed := [crypto.md5(regex.replace(text, `\s+`, "")) | some text in rules_as_text]

	some possible_duplicates in util.find_duplicates(rules_hashed)

	# need to include the original index here to be able to backtrack that to the rule
	asts := {index: ast |
		some index in possible_duplicates

		module := sprintf("package p\n\nimport rego.v1\n\n%s", [rules_as_text[index]])

		# note that we _don't_ use regal.parse_module here, as we do not want location
		# information — only the structure of the AST must match
		ast := rego.parse_module("", module)
	}

	keys := [key | some key, _ in asts]
	vals := [val | some val in asts]

	indices := [keys[index] |
		some dups in util.find_duplicates(vals)
		some index in dups
	]

	count(indices) > 0
}
