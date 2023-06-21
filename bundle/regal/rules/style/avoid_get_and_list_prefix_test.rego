package regal.rules.style_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style
import data.regal.rules.style.common_test.report

test_fail_rule_name_starts_with_get if {
	r := report(`get_foo := 1`)
	r == {{
		"category": "style",
		"description": "Avoid get_ and list_ prefix for rules and functions",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/avoid-get-and-list-prefix", "style"),
		}],
		"title": "avoid-get-and-list-prefix",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": "get_foo := 1"},
		"level": "error",
	}}
}

test_fail_function_name_starts_with_list if {
	r := report(`list_users(datasource) := ["we", "have", "no", "users"]`)
	r == {{
		"category": "style",
		"description": "Avoid get_ and list_ prefix for rules and functions",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/avoid-get-and-list-prefix", "style"),
		}],
		"title": "avoid-get-and-list-prefix",
		"location": {
			"col": 1, "file": "policy.rego", "row": 8,
			"text": `list_users(datasource) := ["we", "have", "no", "users"]`,
		},
		"level": "error",
	}}
}
