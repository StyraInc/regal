package regal.rules.bugs["if-object-literal_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["if-object-literal"] as rule

test_fail_if_empty_object if {
	module := ast.with_rego_v1("rule if {}")
	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Object literal following `if`",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 5, "text": "rule if {}"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/if-object-literal", "bugs"),
		}],
		"title": "if-object-literal",
	}}
}

test_fail_non_empty_object if {
	module := ast.with_rego_v1(`rule if {"x": input.x}`)
	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Object literal following `if`",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 5, "text": `rule if {"x": input.x}`},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/if-object-literal", "bugs"),
		}],
		"title": "if-object-literal",
	}}
}

test_success_not_an_object if {
	module := ast.with_rego_v1(`rule if { true }`)
	r := rule.report with input as module
	r == set()
}
