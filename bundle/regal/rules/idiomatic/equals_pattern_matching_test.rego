package regal.rules.idiomatic["equals-pattern-matching_test"]

import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config

import data.regal.rules.idiomatic["equals-pattern-matching"] as rule

test_fail_simple_head_comparison_could_be_matched_in_arg if {
	module := ast.policy("f(x) := x == 1")

	r := rule.report with input as module
	r == expected_with_location({"col": 1, "file": "policy.rego", "row": 3, "text": "f(x) := x == 1"})
}

test_fail_simple_head_comparison_could_be_matched_in_arg_eq_order if {
	module := ast.policy("f(x) := 1 == x")

	r := rule.report with input as module
	r == expected_with_location({"col": 1, "file": "policy.rego", "row": 3, "text": "f(x) := 1 == x"})
}

test_fail_simple_head_comparison_could_be_matched_in_arg_multiple_args if {
	module := ast.policy("f(_, x, _) := x == 1")

	r := rule.report with input as module
	r == expected_with_location({"col": 1, "file": "policy.rego", "row": 3, "text": "f(_, x, _) := x == 1"})
}

test_fail_simple_body_comparison_could_be_matched_in_arg if {
	module := ast.policy(`f(x) := "one" {
		x == 1
	}`)

	r := rule.report with input as module
	r == expected_with_location({"col": 1, "file": "policy.rego", "row": 3, "text": "f(x) := \"one\" {"})
}

test_fail_simple_body_comparison_could_be_matched_in_arg_eq_order if {
	module := ast.policy(`f(x) := "one" {
		1 == x
	}`)

	r := rule.report with input as module
	r == expected_with_location({"col": 1, "file": "policy.rego", "row": 3, "text": "f(x) := \"one\" {"})
}

test_fail_simple_body_comparison_could_be_matched_using_if if {
	module := ast.with_future_keywords(`f(x) := x if x == 1`)

	r := rule.report with input as module
	r == expected_with_location({"col": 1, "file": "policy.rego", "row": 8, "text": "f(x) := x if x == 1"})
}

test_success_actually_pattern_matching if {
	module := ast.policy("f(1)")

	r := rule.report with input as module
	r == set()
}

test_success_skipped_on_else if {
	module := ast.policy(`f(x) {
		x == 1
	} else := false`)

	r := rule.report with input as module
	r == set()
}

expected := {
	"category": "idiomatic",
	"description": "Prefer pattern matching in function arguments",
	"level": "error",
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/equals-pattern-matching", "idiomatic"),
	}],
	"title": "equals-pattern-matching",
}

expected_with_location(location) := {object.union(expected, {"location": location})} if is_object(location)

expected_with_location(locations) := {object.union(expected, {"location": location}) |
	some location in locations
} if {
	is_set(locations)
}
