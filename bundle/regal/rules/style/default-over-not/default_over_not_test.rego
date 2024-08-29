package regal.rules.style["default-over-not_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.style["default-over-not"] as rule

test_fail_default_over_not if {
	module := ast.with_rego_v1(`
	user := input.user
	user := "foo" if not input.user
	`)
	r := rule.report with input as module
	r == {{
		"category": "style",
		"description": "Prefer default assignment over negated condition",
		"level": "error",
		"location": {"col": 19, "file": "policy.rego", "row": 7, "text": "\tuser := \"foo\" if not input.user"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/default-over-not", "style"),
		}],
		"title": "default-over-not",
	}}
}

test_success_non_constant_value if {
	module := ast.with_rego_v1(`
	user := input.user
	user := var if not input.user
	`)
	r := rule.report with input as module
	r == set()
}

test_success_var_in_ref if {
	module := ast.with_rego_v1(`
	user := input[x].user
	user := "foo" if not input[x].user
	`)
	r := rule.report with input as module
	r == set()
}
