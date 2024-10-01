package regal.rules.idiomatic["boolean-assignment_test"]

import rego.v1

import data.regal.ast
import data.regal.capabilities
import data.regal.config

import data.regal.rules.idiomatic["boolean-assignment"] as rule

test_boolean_assignment_in_rule_head if {
	r := rule.report with input as ast.policy("more_than_one := input.count > 1")
		with data.internal.combined_config as {"capabilities": capabilities.provided}

	r == {{
		"category": "idiomatic",
		"description": "Prefer `if` over boolean assignment",
		"level": "error",
		"location": {
			"col": 1,
			"file": "policy.rego",
			"row": 3,
			"end": {
				"col": 33,
				"row": 3,
			},
			"text": "more_than_one := input.count > 1",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/boolean-assignment", "idiomatic"),
		}],
		"title": "boolean-assignment",
	}}
}

test_success_uses_if_instead_of_boolean_assignment_in_rule_head if {
	r := rule.report with input as ast.with_rego_v1("more_than_one if input.count > 1")
		with data.internal.combined_config as {"capabilities": capabilities.provided}

	r == set()
}

test_success_non_boolean_assignment_in_rule_head if {
	r := rule.report with input as ast.with_rego_v1(`foo := lower("FOO")`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}

	r == set()
}
