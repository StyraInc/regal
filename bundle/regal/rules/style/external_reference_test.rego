package regal.rules.style["external-reference_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style["external-reference"] as rule

test_fail_function_references_input if {
	r := rule.report with input as ast.policy(`f(_) { input.foo }`)
	r == {{
		"category": "style",
		"description": "Reference to input, data or rule ref in function body",
		"location": {"col": 8, "file": "policy.rego", "row": 3, "text": `f(_) { input.foo }`},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "style"),
		}],
		"title": "external-reference",
		"level": "error",
	}}
}

test_fail_function_references_data if {
	r := rule.report with input as ast.policy(`f(_) { data.foo }`)
	r == {{
		"category": "style",
		"description": "Reference to input, data or rule ref in function body",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "style"),
		}],
		"title": "external-reference",
		"location": {"col": 8, "file": "policy.rego", "row": 3, "text": `f(_) { data.foo }`},
		"level": "error",
	}}
}

test_fail_function_references_rule if {
	r := rule.report with input as ast.policy(`
foo := "bar"

f(x, y) {
	x == 5
	y == foo
}
	`)
	r == {{
		"category": "style",
		"description": "Reference to input, data or rule ref in function body",
		"location": {"col": 7, "file": "policy.rego", "row": 8, "text": `	y == foo`},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "style"),
		}],
		"title": "external-reference",
		"level": "error",
	}}
}

test_success_function_references_no_input_or_data if {
	r := rule.report with input as ast.policy(`f(x) { x == true }`)
	r == set()
}

test_success_function_references_no_input_or_data_reverse if {
	r := rule.report with input as ast.policy(`f(x) { true == x }`)
	r == set()
}

test_success_function_references_only_own_vars if {
	r := rule.report with input as ast.policy(`f(x) { y := x; y == 10 }`)
	r == set()
}

test_success_function_references_only_own_vars_nested if {
	r := rule.report with input as ast.policy(`f(x, z) { y := x; y == [1, 2, z]}`)
	r == set()
}
