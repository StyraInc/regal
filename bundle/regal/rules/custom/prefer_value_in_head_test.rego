package regal.rules.custom["prefer-value-in-head_test"]

import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config

import data.regal.rules.custom["prefer-value-in-head"] as rule

test_fail_value_could_be_in_head_assign if {
	module := ast.policy(`value := x {
		input.x
		x := 10
	}`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == expected_with_location({"col": 3, "file": "policy.rego", "row": 5, "text": "\t\tx := 10"})
}

test_fail_value_could_be_in_head_assign_composite if {
	module := ast.policy(`value := x {
		input.x
		x := {"foo": 10}
	}`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == expected_with_location({"col": 3, "file": "policy.rego", "row": 5, "text": "\t\tx := {\"foo\": 10}"})
}

test_fail_value_is_in_head_assign if {
	module := ast.policy(`value := 10 { input.x }`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == set()
}

test_fail_value_could_be_in_head_eq if {
	module := ast.policy(`value := x {
		input.x
		x = 10
	}`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == expected_with_location({"col": 3, "file": "policy.rego", "row": 5, "text": "\t\tx = 10"})
}

test_success_value_is_in_head_eq if {
	module := ast.policy(`value = x { input.x }`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == set()
}

test_fail_value_could_be_in_head_but_not_a_required_scalar if {
	module := ast.policy(`value := x {
		input.x
		x := [i | i := input[_]]
	}`)

	r := rule.report with input as module with config.for_rule as {"level": "error", "only-scalars": true}
	r == expected_with_location({"col": 3, "file": "policy.rego", "row": 5, "text": "\t\tx := [i | i := input[_]]"})
}

test_success_value_could_be_in_head_and_is_a_required_scalar if {
	module := ast.policy(`value := x {
		input.x
		x := 5
	}`)

	r := rule.report with input as module with config.for_rule as {"level": "error", "only-scalars": true}
	r == set()
}

test_fail_value_could_be_in_head_multivalue_rule if {
	module := ast.with_future_keywords(`violations contains violation if {
		input.bad
		violation := "not good!"
	}`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == expected_with_location({"col": 3, "file": "policy.rego", "row": 10, "text": "\t\tviolation := \"not good!\""})
}

test_fail_value_could_be_in_head_object_rule if {
	module := ast.policy(`foo["bar"] := x {
		input.foo
		x := "bar"
	}`)

	r := rule.report with input as module with config.for_rule as {"level": "error"}
	r == expected_with_location({"col": 3, "file": "policy.rego", "row": 5, "text": "\t\tx := \"bar\""})
}

expected := {
	"category": "custom",
	"description": "Prefer value in rule head",
	"level": "error",
	"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "foo := true"},
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/prefer-value-in-head", "custom"),
	}],
	"title": "prefer-value-in-head",
}

expected_with_location(location) := {object.union(expected, {"location": location})}
