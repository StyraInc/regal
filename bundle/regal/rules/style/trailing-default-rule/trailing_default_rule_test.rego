package regal.rules.style["trailing-default-rule_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.style["trailing-default-rule"] as rule

test_success_default_declared_first if {
	module := ast.with_rego_v1(`
	default foo := true

	foo if true
	`)
	r := rule.report with input as module

	r == set()
}

test_fail_default_declared_after if {
	module := ast.with_rego_v1(`
	foo if true

	default foo := true
	`)
	r := rule.report with input as module

	r == {{
		"category": "style",
		"description": "Default rule should be declared first",
		"level": "error",
		"location": {
			"col": 2,
			"file": "policy.rego",
			"row": 8,
			"end": {
				"col": 9,
				"row": 8,
			},
			"text": "\tdefault foo := true",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/trailing-default-rule", "style"),
		}],
		"title": "trailing-default-rule",
	}}
}
