package regal.rules.bugs["rule-assigns-default_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["rule-assigns-default"] as rule

test_fail_rule_assigned_default_value if {
	module := ast.with_rego_v1(`

	default allow := false

	allow := false if {
		some conditions in policy
	}
	`)
	r := rule.report with input as module

	r == {{
		"category": "bugs",
		"description": "Rule assigned its default value",
		"level": "error",
		"location": {
			"col": 11,
			"end": {
				"col": 16,
				"row": 9,
			},
			"file": "policy.rego",
			"row": 9,
			"text": "\tallow := false if {",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/rule-assigns-default", "bugs"),
		}],
		"title": "rule-assigns-default",
	}}
}

test_success_rule_not_assigned_default_value if {
	module := ast.with_rego_v1(`

	default allow := false

	allow := true if {
		some conditions in policy
	}
	`)
	r := rule.report with input as module

	r == set()
}
