package regal.rules.style["unnecessary-some_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.style["unnecessary-some"] as rule

test_fail_some_unnecessary_value if {
	module := ast.with_rego_v1(`
	rule if {
		some "x" in ["x"]
	}
	`)
	r := rule.report with input as module

	r == {{
		"category": "style",
		"description": "Unnecessary use of `some`",
		"level": "error",
		"location": {
			"col": 8,
			"file": "policy.rego",
			"row": 7,
			"end": {
				"col": 20,
				"row": 7,
			},
			"text": "\t\tsome \"x\" in [\"x\"]",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/unnecessary-some", "style"),
		}],
		"title": "unnecessary-some",
	}}
}

test_fail_some_unnecessary_key_value if {
	module := ast.with_rego_v1(`
	rule if {
		some "x", 1 in {"x": 1}
	}
	`)
	r := rule.report with input as module

	r == {{
		"category": "style",
		"description": "Unnecessary use of `some`",
		"level": "error",
		"location": {
			"col": 8,
			"file": "policy.rego",
			"row": 7,
			"end": {
				"col": 26,
				"row": 7,
			},
			"text": "\t\tsome \"x\", 1 in {\"x\": 1}",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/unnecessary-some", "style"),
		}],
		"title": "unnecessary-some",
	}}
}

test_success_some_value_using_var if {
	module := ast.with_rego_v1(`
	rule if {
		some var in input.vars
	}
	`)
	r := rule.report with input as module

	r == set()
}

test_success_some_key_value_using_var_for_value if {
	module := ast.with_rego_v1(`
	rule if {
		some "x", var in {"x": 1}
	}
	`)
	r := rule.report with input as module

	r == set()
}

test_success_some_key_value_using_var_for_key if {
	module := ast.with_rego_v1(`
	rule if {
		some var, 1 in {"x": 1}
	}
	`)
	r := rule.report with input as module

	r == set()
}

test_success_just_in_head if {
	r := rule.report with input as ast.with_rego_v1(`rule := [1 in []]`)

	r == set()
}
