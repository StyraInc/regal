package regal.rules.bugs["argument-always-wildcard_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["argument-always-wildcard"] as rule

test_fail_single_function_single_argument_always_a_wildcard if {
	module := ast.with_rego_v1(`
	f(_) := 1
	`)

	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Argument is always a wildcard",
		"level": "error",
		"location": {"col": 4, "file": "policy.rego", "row": 6, "text": "\tf(_) := 1"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/argument-always-wildcard", "bugs"),
		}],
		"title": "argument-always-wildcard",
	}}
}

test_fail_single_argument_always_a_wildcard if {
	module := ast.with_rego_v1(`
	f(_) := 1
	f(_) := 2
	`)

	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Argument is always a wildcard",
		"level": "error",
		"location": {"col": 4, "file": "policy.rego", "row": 6, "text": "\tf(_) := 1"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/argument-always-wildcard", "bugs"),
		}],
		"title": "argument-always-wildcard",
	}}
}

test_fail_single_argument_always_a_wildcard_default_function if {
	module := ast.with_rego_v1(`
	default f(_) := 1
	f(_) := 2
	`)

	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Argument is always a wildcard",
		"level": "error",
		"location": {"col": 12, "file": "policy.rego", "row": 6, "text": "\tdefault f(_) := 1"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/argument-always-wildcard", "bugs"),
		}],
		"title": "argument-always-wildcard",
	}}
}

test_fail_multiple_argument_always_a_wildcard if {
	module := ast.with_rego_v1(`
	f(x, _) := x + 1
	f(x, _) := x + 2
	`)

	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Argument is always a wildcard",
		"level": "error",
		"location": {"col": 7, "file": "policy.rego", "row": 6, "text": "\tf(x, _) := x + 1"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/argument-always-wildcard", "bugs"),
		}],
		"title": "argument-always-wildcard",
	}}
}

test_success_multiple_argument_not_always_a_wildcard if {
	module := ast.with_rego_v1(`
	f(x, _) := x + 1
	f(_, y) := y + 2
	`)

	r := rule.report with input as module
	r == set()
}
