package regal.rules.bugs["redundant-existence-check_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["redundant-existence-check"] as rule

test_fail_redundant_existence_check if {
	module := ast.with_rego_v1(`
	redundant if {
		input.foo
		startswith(input.foo, "bar")
	}`)
	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Redundant existence check",
		"level": "error",
		"location": {"col": 3, "file": "policy.rego", "row": 7, "text": "\t\tinput.foo", "end": {"col": 12, "row": 7}},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/redundant-existence-check", "bugs"),
		}],
		"title": "redundant-existence-check",
	}}
}

test_success_not_redundant_existence_check if {
	module := ast.with_rego_v1(`
	redundant if {
		input.foo
		something_expensive
		startswith(input.foo, "bar")
	}`)
	r := rule.report with input as module
	r == set()
}

test_success_not_redundant_existence_check_with_cancels if {
	module := ast.with_rego_v1(`
	not_redundant if {
		rule.foo with input as {}
		rule.foo == 1
	}`)
	r := rule.report with input as module
	r == set()
}

test_fail_redundant_existence_check_head_assignment_of_ref if {
	module := ast.with_rego_v1(`
	redundant := input.foo if {
		input.foo
	}`)
	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Redundant existence check",
		"level": "error",
		"location": {"col": 3, "file": "policy.rego", "row": 7, "text": "\t\tinput.foo", "end": {"col": 12, "row": 7}},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/redundant-existence-check", "bugs"),
		}],
		"title": "redundant-existence-check",
	}}
}

test_fail_redundant_existence_check_function_arg if {
	module := ast.with_rego_v1(`
	fun(foo) if {
		foo
	}`)
	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Redundant existence check",
		"level": "error",
		"location": {"col": 3, "end": {"col": 6, "row": 7}, "file": "policy.rego", "row": 7, "text": "\t\tfoo"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/redundant-existence-check", "bugs"),
		}],
		"title": "redundant-existence-check",
	}}
}
