package regal.rules.idiomatic["use-if_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.idiomatic["use-if"] as rule

test_fail_should_use_if if {
	module := ast.with_rego_v0(`rule := [true |
		input[_]
	] {
		input.attribute
	}`)
	r := rule.report with input as module

	r == {{
		"category": "idiomatic",
		"description": "Use the `if` keyword",
		"level": "error",
		"location": {
			"col": 1,
			"file": "policy_v0.rego",
			"row": 3,
			"end": {
				"col": 5,
				"row": 3,
			},
			"text": "rule := [true |",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-if", "idiomatic"),
		}],
		"title": "use-if",
	}}
}

test_success_uses_if if {
	module := ast.with_rego_v1(`rule := [true |
		input[_]
	] if {
		input.attribute
	}`)
	r := rule.report with input as module

	r == set()
}

test_success_no_body_no_if if {
	r := rule.report with input as ast.with_rego_v1(`rule := "without body"`)

	r == set()
}
