package regal.rules.idiomatic["use-contains_test"]

import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config

import data.regal.rules.idiomatic["use-contains"] as rule

test_fail_should_use_contains if {
	module := ast.with_future_keywords(`rule[item] {
		some item in input.items
	}`)

	r := rule.report with input as module
	r == {{
		"category": "idiomatic",
		"description": "Use the `contains` keyword",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": "rule[item] {"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-contains", "idiomatic"),
		}],
		"title": "use-contains",
	}}
}

test_success_uses_contains if {
	module := ast.with_future_keywords(`rule contains item if {
		some item in input.items
	}`)

	r := rule.report with input as module
	r == set()
}

test_success_object_rule if {
	module := ast.with_future_keywords(`rule[foo] := bar if {
		foo := "bar"
		bar := "baz"
	}`)

	r := rule.report with input as module
	r == set()
}
