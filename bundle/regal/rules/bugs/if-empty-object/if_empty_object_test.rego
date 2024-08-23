package regal.rules.bugs["if-empty-object_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["if-empty-object"] as rule

test_fail_if_empty_object if {
	module := ast.with_rego_v1("rule if {}")
	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Empty object following `if`",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 5, "text": "rule if {}"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/if-empty-object", "bugs"),
		}],
		"title": "if-empty-object",
	}}
}

test_success_if_non_empty_object if {
	# this is arguably just as useless, but we'll defer
	# to the constant-condition rule for these cases
	module := ast.with_rego_v1(`rule if {"foo": "bar"}`)
	r := rule.report with input as module
	r == set()
}
