package regal.rules.bugs["not-equals-in-loop_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.bugs["not-equals-in-loop"] as rule

test_fail_neq_in_loop if {
	r := rule.report with input as ast.policy(`deny {
		"admin" != input.user.groups[_]
		input.user.groups[_] != "admin"
	}`)
	r == {
		{
			"category": "bugs",
			"description": "Use of != in loop",
			"level": "error",
			"location": {"col": 11, "file": "policy.rego", "row": 4, "text": "\t\t\"admin\" != input.user.groups[_]"},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/not-equals-in-loop", "bugs"),
			}],
			"title": "not-equals-in-loop",
		},
		{
			"category": "bugs",
			"description": "Use of != in loop",
			"level": "error",
			"location": {"col": 24, "file": "policy.rego", "row": 5, "text": "\t\tinput.user.groups[_] != \"admin\""},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/not-equals-in-loop", "bugs"),
			}],
			"title": "not-equals-in-loop",
		},
	}
}

test_fail_neq_in_loop_one_liner if {
	r := rule.report with input as ast.with_future_keywords(`deny if "admin" != input.user.groups[_]`)
	r == {{
		"category": "bugs",
		"description": "Use of != in loop",
		"level": "error",
		"location": {"col": 17, "file": "policy.rego", "row": 8, "text": "deny if \"admin\" != input.user.groups[_]"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/not-equals-in-loop", "bugs"),
		}],
		"title": "not-equals-in-loop",
	}}
}
