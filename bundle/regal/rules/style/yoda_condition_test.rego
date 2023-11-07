package regal.rules.style["yoda-condition_test"]

import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config

import data.regal.rules.style["yoda-condition"] as rule

test_fail_yoda_conditions if {
	module := ast.policy(`rule {
		"foo" == input.bar

		nested := [foo |
			foo := input.baz
			"foo" == foo
		]
	}`)
	r := rule.report with input as module
	r == expected_with_location([
		{"col": 9, "file": "policy.rego", "row": 4, "text": "\t\t\"foo\" == input.bar"},
		{"col": 10, "file": "policy.rego", "row": 8, "text": "\t\t\t\"foo\" == foo"},
	])
}

test_fail_yoda_conditions_not_equals if {
	module := ast.policy(`rule {
		"foo" != input.bar

		nested := [foo |
			foo := input.baz
			"foo" != foo
		]
	}`)
	r := rule.report with input as module
	r == expected_with_location([
		{"col": 9, "file": "policy.rego", "row": 4, "text": "\t\t\"foo\" != input.bar"},
		{"col": 10, "file": "policy.rego", "row": 8, "text": "\t\t\t\"foo\" != foo"},
	])
}

test_success_no_yoda_condition if {
	module := ast.policy(`rule {
		input.bar == "foo"
	}`)
	r := rule.report with input as module
	r == set()
}

test_success_constants_on_both_sides if {
	module := ast.policy(`rule {
		"foo" == "foo"
	}`)
	r := rule.report with input as module
	r == set()
}

test_success_exclude_ref_with_vars if {
	module := ast.policy(`rule {
		"foo" == input.bar[_]
	}`)
	r := rule.report with input as module
	r == set()
}

expected := {
	"category": "style",
	"description": "Yoda condition",
	"level": "error",
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/yoda-condition", "style"),
	}],
	"title": "yoda-condition",
}

expected_with_location(location) := {object.union(expected, {"location": location})} if is_object(location)

expected_with_location(location) := {object.union(expected, {"location": loc}) |
	some loc in location
} if {
	is_array(location)
}
