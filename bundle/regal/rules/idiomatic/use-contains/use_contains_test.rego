package regal.rules.idiomatic["use-contains_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.idiomatic["use-contains"] as rule

test_fail_should_use_contains if {
	module := ast.policy(`
	import future.keywords

	rule[item] {
		some item in input.items
	}`)
	r := rule.report with input as module

	r == {{
		"category": "idiomatic",
		"description": "Use the `contains` keyword",
		"level": "error",
		"location": {
			"col": 2,
			"file": "policy.rego",
			"row": 6,
			"end": {
				"row": 6,
				"col": 12,
			},
			"text": "\trule[item] {",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-contains", "idiomatic"),
		}],
		"title": "use-contains",
	}}
}

test_success_uses_contains if {
	module := ast.with_rego_v1(`rule contains item if {
		some item in input.items
	}`)
	r := rule.report with input as module

	r == set()
}

test_success_object_rule if {
	module := ast.with_rego_v1(`rule[foo] := bar if {
		foo := "bar"
		bar := "baz"
	}`)
	r := rule.report with input as module

	r == set()
}
