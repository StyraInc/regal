package regal.rules.style_test

import future.keywords.if

import data.regal.config
import data.regal.rules.style.common_test.report

test_fail_function_references_input if {
	report(`f(_) { input.foo }`) == {{
		"category": "style",
		"description": "Reference to input, data or rule ref in function body",
		"location": {"col": 8, "file": "policy.rego", "row": 8, "text": `f(_) { input.foo }`},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "style"),
		}],
		"title": "external-reference",
		"level": "error",
	}}
}

test_fail_function_references_data if {
	report(`f(_) { data.foo }`) == {{
		"category": "style",
		"description": "Reference to input, data or rule ref in function body",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "style"),
		}],
		"title": "external-reference",
		"location": {"col": 8, "file": "policy.rego", "row": 8, "text": `f(_) { data.foo }`},
		"level": "error",
	}}
}

test_fail_function_references_rule if {
	r := report(`
foo := "bar"

f(x, y) {
	x == 5
	y == foo
}
	`)
	r == {{
		"category": "style",
		"description": "Reference to input, data or rule ref in function body",
		"location": {"col": 7, "file": "policy.rego", "row": 13, "text": `	y == foo`},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "style"),
		}],
		"title": "external-reference",
		"level": "error",
	}}
}

test_success_function_references_no_input_or_data if {
	report(`f(x) { x == true }`) == set()
}

test_success_function_references_no_input_or_data_reverse if {
	report(`f(x) { true == x }`) == set()
}

test_success_function_references_only_own_vars if {
	report(`f(x) { y := x; y == 10 }`) == set()
}

test_success_function_references_only_own_vars_nested if {
	report(`f(x, z) { y := x; y == [1, 2, z]}`) == set()
}
