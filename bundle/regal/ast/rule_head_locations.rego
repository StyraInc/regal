package regal.ast

import rego.v1

import data.regal.util

# METADATA
# description: |
#   For a given rule head name, this rule contains a list of locations where
#   there is a rule head with that name.
rule_head_locations[name] contains {"row": loc.row, "col": loc.col} if {
	some rule in input.rules

	name := concat(".", [
		"data",
		package_name,
		ref_static_to_string(rule.head.ref),
	])

	loc := util.to_location_object(rule.head.location)
}
