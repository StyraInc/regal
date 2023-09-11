package regal.rules.style["chained-rule-body_test"]

import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config

import data.regal.rules.style["chained-rule-body"] as rule

test_fail_chained_incremental_definition if {
	module := ast.policy(`rule {
		input.x
	} {
		input.y
	}`)

	r := rule.report with input as module

	r == {{
		"category": "style",
		"description": "Avoid chaining rule bodies",
		"level": "error",
		"location": {"col": 4, "file": "policy.rego", "row": 5, "text": "\t} {"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/chained-rule-body", "style"),
		}],
		"title": "chained-rule-body",
	}}
}

test_success_not_chained_incremental_definition if {
	module := ast.policy(`
	rule {
		input.x
	}

	rule {
		input.y
	}`)

	r := rule.report with input as module
	r == set()
}
