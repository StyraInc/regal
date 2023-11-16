package regal.rules.idiomatic["custom-in-construct_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.idiomatic["custom-in-construct"] as rule

test_fail_custom_in if {
	r := rule.report with input as ast.policy(`has(item, coll) {
    		item == coll[_]
    }`)
	r == {{
		"category": "idiomatic",
		"description": "Custom function may be replaced by `in` keyword",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "has(item, coll) {"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/custom-in-construct", "idiomatic"),
		}],
		"title": "custom-in-construct",
	}}
}

test_fail_custom_in_reversed if {
	r := rule.report with input as ast.policy(`has(item, coll) { coll[_] == item }`)
	r == {{
		"category": "idiomatic",
		"description": "Custom function may be replaced by `in` keyword",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "has(item, coll) { coll[_] == item }"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/custom-in-construct", "idiomatic"),
		}],
		"title": "custom-in-construct",
	}}
}
