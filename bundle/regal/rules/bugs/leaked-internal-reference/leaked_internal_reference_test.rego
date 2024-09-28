package regal.rules.bugs["leaked-internal-reference_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["leaked-internal-reference"] as rule

test_fail_leaked_internal_reference_in_import if {
	r := rule.report with input as ast.with_rego_v1(`import data.foo._bar`)

	r == expected_with_location({
		"col": 8,
		"row": 5,
		"end": {
			"col": 21,
			"row": 5,
		},
		"text": "import data.foo._bar",
	})
}

test_fail_leaked_internal_reference_in_rule_head if {
	r := rule.report with input as ast.with_rego_v1(`var := data.foo._bar`)

	r == expected_with_location({
		"col": 8,
		"file": "policy.rego",
		"row": 5,
		"end": {
			"col": 21,
			"row": 5,
		},
		"text": "var := data.foo._bar",
	})
}

test_fail_leaked_internal_reference_in_rule_body if {
	r := rule.report with input as ast.with_rego_v1(`rule if {
		x := data.foo._bar
	}`)

	r == expected_with_location({
		"col": 8,
		"row": 6,
		"end": {
			"col": 21,
			"row": 6,
		},
		"text": "\t\tx := data.foo._bar",
	})
}

test_fail_leaked_internal_reference_in_nested_comprehension if {
	r := rule.report with input as ast.with_rego_v1(`rule if {
		comp := [x | x := data.foo._bar]
	}`)

	r == expected_with_location({
		"col": 21,
		"row": 6,
		"end": {
			"col": 34,
			"row": 6,
		},
		"text": "\t\tcomp := [x | x := data.foo._bar]",
	})
}

expected := {
	"category": "bugs",
	"description": "Outside reference to internal rule or function",
	"level": "error",
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/leaked-internal-reference", "bugs"),
	}],
	"title": "leaked-internal-reference",
	"location": {"file": "policy.rego"},
}

# regal ignore:external-reference
expected_with_location(location) := {object.union(expected, {"location": location})} if is_object(location)
