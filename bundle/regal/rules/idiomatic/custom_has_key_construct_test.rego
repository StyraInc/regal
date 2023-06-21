package regal.rules.idiomatic_test

import future.keywords.if

import data.regal.config
import data.regal.rules.idiomatic.common_test.report

test_fail_unnecessary_construct_in if {
	r := report(`has_key(name, coll) {
		_ = coll[name]
	}`)
	r == {{
		"category": "idiomatic",
		"description": "Custom function may be replaced by `in` and `object.keys`",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "has_key(name, coll) {"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/custom-has-key-construct", "idiomatic"),
		}],
		"title": "custom-has-key-construct",
	}}
}

test_fail_unnecessary_construct_in_reversed if {
	r := report(`has_key(name, coll) {
		coll[name] = _
	}`)
	r == {{
		"category": "idiomatic",
		"description": "Custom function may be replaced by `in` and `object.keys`",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "has_key(name, coll) {"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/custom-has-key-construct", "idiomatic"),
		}],
		"title": "custom-has-key-construct",
	}}
}
