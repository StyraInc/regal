package regal.rules.style["external-reference_test"]

import rego.v1

import data.regal.ast
import data.regal.capabilities
import data.regal.config
import data.regal.rules.style["external-reference"] as rule

test_fail_function_references_input if {
	r := rule.report with input as ast.policy(`f(_) { input.foo }`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
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
		with data.internal.combined_config as {"capabilities": capabilities.provided}
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

test_fail_function_references_data_in_expr if {
	r := rule.report with input as ast.policy(`f(x) {
		x == data.foo
	}`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == {{
		"category": "style",
		"description": "Reference to input, data or rule ref in function body",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "style"),
		}],
		"title": "external-reference",
		"location": {"col": 8, "file": "policy.rego", "row": 4, "text": "\t\tx == data.foo"},
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
		with data.internal.combined_config as {"capabilities": capabilities.provided}
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
