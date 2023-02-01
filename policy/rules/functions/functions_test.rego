package regal.rules.functions_test

import future.keywords.if

import data.regal
import data.regal.rules.functions

test_fail_function_references_input if {
	ast := regal.ast(`f(_) { input.foo }`)
	result := functions.violation with input as ast
	result == {{
		"description": "Reference to input or data in function body",
		"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-functions-001"}],
		"title": "STY-FUNCTIONS-001",
	}}
}

test_fail_function_references_data if {
	ast := regal.ast(`f(_) { data.foo }`)
	result := functions.violation with input as ast
	result == {{
		"description": "Reference to input or data in function body",
		"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-functions-001"}],
		"title": "STY-FUNCTIONS-001",
	}}
}

test_success_function_references_no_input_or_data if {
	ast := regal.ast(`f(x) { x == true }`)
	result := functions.violation with input as ast
	result == set()
}
