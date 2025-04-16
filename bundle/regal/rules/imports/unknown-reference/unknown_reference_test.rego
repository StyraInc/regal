package regal.rules.imports["unknown-reference_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.imports["unknown-reference"] as rule

test_fail_identifies_unknown_reference if {
	agg1 := rule.aggregate with input as regal.parse_module("p1.rego", `package foo
	import data.bar

	x := bar.baz
	`)
	agg2 := rule.aggregate with input as regal.parse_module("p2.rego", `package bar
	import data.foo

	x := 1
	`)
	r := rule.aggregate_report with input as {"aggregate": (agg1 | agg2)}
	r == {expected_report({
		"file": "p1.rego",
		"row": 4,
		"col": 7,
		"end": {
			"row": 4,
			"col": 10,
		},
		"text": "\timport data.nope",
		"reference": "bar.baz",
		"full_name": "data.bar.baz",
	})}
}

test_success_no_unknown_reference if {
	agg1 := rule.aggregate with input as regal.parse_module("p1.rego", `package foo
	import data.bar

	x := bar.baz
	`)
	agg2 := rule.aggregate with input as regal.parse_module("p2.rego", `package bar
	import data.foo

	baz := 1
	`)
	r := rule.aggregate_report with input as {"aggregate": (agg1 | agg2)}
	r == set()
}

expected_report(expected_values) := {
	"category": "imports",
	"description": "Reference to unknown field.",
	"level": "error",
	"location": {
		"file": to_location(expected_values),
		"text": sprintf("%s (%s) is unknown", [expected_values.reference, expected_values.full_name]),
	},
	# "location": location,
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/unknown-reference", "imports"),
	}],
	"title": "unknown-reference",
}

# TODO: this should happen more automatically once we have report.location working
to_location(location) := sprintf("%s:%v:%v:%v:%v", [
	location.file,
	location.row,
	location.col,
	location.end.row,
	location.end.col,
])
