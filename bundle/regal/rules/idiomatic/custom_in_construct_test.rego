package regal.rules.idiomatic_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.idiomatic
import data.regal.rules.idiomatic.common_test.report

test_fail_unnecessary_construct_in if {
	r := report(`has(item, coll) {
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

test_fail_unnecessary_construct_in_reversed if {
	r := report(`has(item, coll) { coll[_] == item }`)
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
