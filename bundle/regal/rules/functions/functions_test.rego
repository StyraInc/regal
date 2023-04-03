package regal.rules.functions_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.functions

test_fail_function_references_input if {
	report(`f(_) { input.foo }`) == {{
		"category": "functions",
		"description": "Reference to input, data or rule ref in function body",
		"location": {"col": 8, "file": "policy.rego", "row": 8},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/input-or-data-reference",
		}],
		"title": "input-or-data-reference",
	}}
}

test_fail_function_references_data if {
	report(`f(_) { data.foo }`) == {{
		"category": "functions",
		"description": "Reference to input, data or rule ref in function body",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/input-or-data-reference",
		}],
		"title": "input-or-data-reference",
		"location": {"col": 8, "file": "policy.rego", "row": 8},
	}}
}

test_fail_function_references_rule if {
	report(`
	foo := "bar"

	f(x, y) {
		x == 5
		y == foo
	}`) == {{
		"category": "functions",
		"description": "Reference to input, data or rule ref in function body",
		"location": {"col": 8, "file": "policy.rego", "row": 13},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/input-or-data-reference",
		}],
		"title": "input-or-data-reference",
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

report(snippet) := report if {
	# regal ignore:input-or-data-reference
	report := functions.report with input as ast.with_future_keywords(snippet)
		with config.for_rule as {"enabled": true}
}
