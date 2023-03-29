package regal.rules.functions_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.functions

test_fail_function_references_input if {
	report(`f(_) { input.foo }`) == {{
		"category": "functions",
		"description": "Reference to input or data in function body",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/input-or-data-reference",
		}],
		"title": "input-or-data-reference",
		"location": {"col": 8, "file": "policy.rego", "row": 8},
	}}
}

test_fail_function_references_data if {
	report(`f(_) { data.foo }`) == {{
		"category": "functions",
		"description": "Reference to input or data in function body",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/input-or-data-reference",
		}],
		"title": "input-or-data-reference",
		"location": {"col": 8, "file": "policy.rego", "row": 8},
	}}
}

test_success_function_references_no_input_or_data if {
	report(`f(x) { x == true }`) == set()
}

report(snippet) := report if {
	report := functions.report with input as ast.with_future_keywords(snippet)
		with config.for_rule as {"enabled": true}
}
