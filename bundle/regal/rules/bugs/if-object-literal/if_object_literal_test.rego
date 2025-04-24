package regal.rules.bugs["if-object-literal_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["if-object-literal"] as rule

test_fail[name] if {
	some name, [policy, location] in {
		"empty_object": [
			`rule if {}`,
			{
				"col": 9,
				"row": 5,
				"text": "rule if {}",
				"end": {
					"col": 11,
					"row": 5,
				},
			},
		],
		"non_empty_object": [
			`rule if {"x": input.x}`,
			{
				"col": 9,
				"row": 5,
				"text": `rule if {"x": input.x}`,
				"end": {
					"col": 23,
					"row": 5,
				},
			},
		],
	}

	r := rule.report with input as ast.with_rego_v1(policy)
	r == {{
		"category": "bugs",
		"description": "Object literal following `if`",
		"level": "error",
		"location": object.union({"file": "policy.rego"}, location),
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
