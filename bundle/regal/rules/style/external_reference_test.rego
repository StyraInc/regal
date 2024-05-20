package regal.rules.style["external-reference_test"]

import rego.v1

import data.regal.ast
import data.regal.capabilities
import data.regal.config
import data.regal.rules.style["external-reference"] as rule

test_fail_function_references_input if {
	r := rule.report with input as ast.policy(`f(_) { input.foo }`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == expected_with_location({"col": 8, "file": "policy.rego", "row": 3, "text": `f(_) { input.foo }`})
}

test_fail_function_references_data if {
	r := rule.report with input as ast.policy(`f(_) { data.foo }`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == expected_with_location({"col": 8, "file": "policy.rego", "row": 3, "text": `f(_) { data.foo }`})
}

test_fail_function_references_data_in_expr if {
	r := rule.report with input as ast.policy(`f(x) {
		x == data.foo
	}`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == expected_with_location({"col": 8, "file": "policy.rego", "row": 4, "text": "\t\tx == data.foo"})
}

test_fail_function_references_rule if {
	r := rule.report with input as ast.policy(`
foo := "bar"

f(x, y) {
	x == 5
	y == foo
}
	`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == expected_with_location({"col": 7, "file": "policy.rego", "row": 8, "text": `	y == foo`})
}

test_fail_external_reference_in_head_assignment if {
	r := rule.report with input as ast.policy(`f(_) := r`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == expected_with_location({"col": 9, "file": "policy.rego", "row": 3, "text": "f(_) := r"})
}

test_fail_external_reference_in_head_terms if {
	r := rule.report with input as ast.policy(`f(_) := {"r": r}`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == expected_with_location({"col": 15, "file": "policy.rego", "row": 3, "text": "f(_) := {\"r\": r}"})
}

test_success_function_references_no_input_or_data if {
	r := rule.report with input as ast.policy(`f(x) { x == true }`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}

test_success_function_references_no_input_or_data_reverse if {
	r := rule.report with input as ast.policy(`f(x) { true == x }`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}

test_success_function_references_only_own_vars if {
	r := rule.report with input as ast.policy(`f(x) { y := x; y == 10 }`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}

test_success_function_references_only_own_vars_nested if {
	r := rule.report with input as ast.policy(`f(x, z) { y := x; y == [1, 2, z]}`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}

test_success_function_references_only_own_vars_and_wildcard if {
	r := rule.report with input as ast.policy(`f(x, y) { _ = x + y }`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}

test_success_function_references_return_var if {
	r := rule.report with input as ast.policy(`f(x) := y { y = true }`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}

test_success_function_references_return_vars if {
	r := rule.report with input as ast.policy(`f(x) := [x, y] { x = false; y = true }`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}

test_success_function_references_external_function if {
	r := rule.report with input as ast.policy(`f(x) { data.foo.bar(x) }`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}

test_success_function_references_external_function_in_expr if {
	r := rule.report with input as ast.policy(`f(x) := y { y := data.foo.bar(x) }`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}

expected := {
	"category": "style",
	"description": "External reference in function",
	"related_resources": [{
		"description": "documentation",
		"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "style"),
	}],
	"title": "external-reference",
	"level": "error",
}

# regal ignore:external-reference
expected_with_location(location) := {object.union(expected, {"location": location})} if is_object(location)
