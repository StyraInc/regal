package regal.rules.functions_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.functions

test_fail_function_references_input if {
	report(`f(_) { input.foo }`) == {{
		"category": "functions",
		"description": "Reference to input, data or rule ref in function body",
		"location": {"col": 8, "file": "policy.rego", "row": 8, "text": `f(_) { input.foo }`},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "functions"),
		}],
		"title": "external-reference",
		"level": "error",
	}}
}

test_fail_function_references_data if {
	report(`f(_) { data.foo }`) == {{
		"category": "functions",
		"description": "Reference to input, data or rule ref in function body",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "functions"),
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
		"category": "functions",
		"description": "Reference to input, data or rule ref in function body",
		"location": {"col": 7, "file": "policy.rego", "row": 13, "text": `	y == foo`},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/external-reference", "functions"),
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

test_fail_call_to_print_and_trace if {
	r := report(`allow {
		print("foo")

		x := [i | i = 0; trace("bar")]
	}`)
	r == {
		{
			"category": "functions",
			"description": "Call to print or trace function",
			"level": "error",
			"location": {"col": 3, "file": "policy.rego", "row": 9, "text": "\t\tprint(\"foo\")"},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/print-or-trace-call", "functions"),
			}],
			"title": "print-or-trace-call",
		},
		{
			"category": "functions",
			"description": "Call to print or trace function",
			"level": "error",
			"location": {"col": 20, "file": "policy.rego", "row": 11, "text": "\t\tx := [i | i = 0; trace(\"bar\")]"},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/print-or-trace-call", "functions"),
			}],
			"title": "print-or-trace-call",
		},
	}
}

report(snippet) := report if {
	# regal ignore:external-reference
	report := functions.report with input as ast.with_future_keywords(snippet)
}
