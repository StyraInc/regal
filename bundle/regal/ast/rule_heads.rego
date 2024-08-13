package regal.ast

import rego.v1

# METADATA
# description: |
#   For a given rule head name, this rule contains a list of locations where
#   there is a rule head with that name.
rule_head_locations[name] contains info if {
	some rule in input.rules

	name := concat(".", [
		"data",
		package_name,
		ref_static_to_string(rule.head.ref),
	])

	info := {
		"row": rule.head.location.row,
		"col": rule.head.location.col,
	}
}
