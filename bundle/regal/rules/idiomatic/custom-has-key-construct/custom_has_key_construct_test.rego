package regal.rules.idiomatic["custom-has-key-construct_test"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.rules.idiomatic["custom-has-key-construct"] as rule

test_fail_custom_has_key if {
	r := rule.report with input as ast.policy(`has_key(name, coll) {
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

test_fail_custom_has_key_reversed if {
	r := rule.report with input as ast.policy(`has_key(name, coll) {
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

test_fail_custom_has_key_multiple_wildcards if {
	r := rule.report with input as ast.policy(`
	other_rule["foo"] {
		wildcard := input[_]
	}

	has_key(name, coll) {
		coll[name] = _
	}`)
	r == {{
		"category": "idiomatic",
		"description": "Custom function may be replaced by `in` and `object.keys`",
		"level": "error",
		"location": {"col": 2, "file": "policy.rego", "row": 8, "text": "\thas_key(name, coll) {"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/custom-has-key-construct", "idiomatic"),
		}],
		"title": "custom-has-key-construct",
	}}
}

test_has_notice_if_unmet_capability if {
	r := rule.notices with config.capabilities as {}
	r == {{
		"category": "idiomatic",
		"description": "Missing capability for built-in function `object.keys`",
		"level": "notice",
		"severity": "warning",
		"title": "custom-has-key-construct",
	}}
}
