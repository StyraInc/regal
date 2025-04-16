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
	expected := {with_location({
		"file": "p1.rego",
		"row": 4,
		"col": 7,
		"end": {
			"row": 4,
			"col": 10,
		},
		"text": "\tx := bar.baz",
	})}
	r == expected
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

test_fail_identifies_unknown_reference_with_alias if {
	agg1 := rule.aggregate with input as regal.parse_module("p1.rego", `package foo
	import data.bar as baz

	x := baz.qux
	`)
	agg2 := rule.aggregate with input as regal.parse_module("p2.rego", `package bar
	import data.foo

	x := 1
	`)
	r := rule.aggregate_report with input as {"aggregate": (agg1 | agg2)}
	expected := {with_location({
		"file": "p1.rego",
		"row": 4,
		"col": 7,
		"end": {
			"row": 4,
			"col": 10,
		},
		"text": "\tx := baz.qux",
	})}
	r == expected
}

test_success_identifies_reference_with_alias if {
	agg1 := rule.aggregate with input as regal.parse_module("p1.rego", `package foo
	import data.bar as baz

	x := baz.qux
	`)
	agg2 := rule.aggregate with input as regal.parse_module("p2.rego", `package bar
	import data.foo

	qux := 1
	`)
	r := rule.aggregate_report with input as {"aggregate": (agg1 | agg2)}
	r == set()
}

test_fail_identifies_unknown_full_path if {
	agg1 := rule.aggregate with input as regal.parse_module("p1.rego", `package foo

	x := data.bar.baz
	`)
	agg2 := rule.aggregate with input as regal.parse_module("p2.rego", `package bar
	import data.foo

	x := 1
	`)
	r := rule.aggregate_report with input as {"aggregate": (agg1 | agg2)}
	expected := {with_location({
		"file": "p1.rego",
		"row": 3,
		"col": 7,
		"end": {
			"row": 3,
			"col": 11,
		},
		"text": "\tx := data.bar.baz",
	})}
	r == expected
}

test_success_identifies_full_path if {
	agg1 := rule.aggregate with input as regal.parse_module("p1.rego", `package foo

	x := data.bar.baz
	`)
	agg2 := rule.aggregate with input as regal.parse_module("p2.rego", `package bar
	import data.foo

	baz := 1
	`)
	r := rule.aggregate_report with input as {"aggregate": (agg1 | agg2)}
	r == set()
}

test_fail_everything_all_at_once if {
	agg1 := rule.aggregate with input as regal.parse_module("p1.rego", `package foo
	import data.bar
	import data.baz as qux

	x := bar.unknown
	y := qux.unknown
	z := data.qux.unknown
	`)
	agg2 := rule.aggregate with input as regal.parse_module("p2.rego", `package bar

	known := 1
	`)
	agg3 := rule.aggregate with input as regal.parse_module("p3.rego", `package baz

	known := 1
	`)
	agg4 := rule.aggregate with input as regal.parse_module("p4.rego", `package qux

	known := 1
	`)
	r := rule.aggregate_report with input as {"aggregate": (agg1 | agg2)}
	expected := {
		with_location({
		"file": "p1.rego",
		"row": 5,
		"col": 7,
		"end": {
			"row": 5,
			"col": 10,
		},
		"text": "\tx := bar.unknown",
	}),
		with_location({
		"file": "p1.rego",
		"row": 6,
		"col": 7,
		"end": {
			"row": 6,
			"col": 10,
		},
		"text": "\ty := qux.unknown",
	}),
		with_location({
		"file": "p1.rego",
		"row": 7,
		"col": 7,
		"end": {
			"row": 7,
			"col": 11,
		},
		"text": "\tz := data.qux.unknown",
	}),
	}
	r == expected
}

with_location(location) := {
	"category": "imports",
	"description": "Reference to unknown field.",
	"level": "error",
	"location": location,
	# "location": location,
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/unknown-reference", "imports"),
	}],
	"title": "unknown-reference",
}
