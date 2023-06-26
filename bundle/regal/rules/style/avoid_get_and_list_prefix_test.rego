package regal.rules.style["avoid-get-and-list-prefix_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style["avoid-get-and-list-prefix"] as rule

test_fail_rule_name_starts_with_get if {
	r := rule.report with input as ast.policy(`get_foo := 1`)
	r == {{
		"category": "style",
		"description": "Avoid get_ and list_ prefix for rules and functions",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/avoid-get-and-list-prefix", "style"),
		}],
		"title": "avoid-get-and-list-prefix",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "get_foo := 1"},
		"level": "error",
	}}
}

test_fail_function_name_starts_with_list if {
	r := rule.report with input as ast.policy(`list_users(datasource) := ["we", "have", "no", "users"]`)
	r == {{
		"category": "style",
		"description": "Avoid get_ and list_ prefix for rules and functions",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/avoid-get-and-list-prefix", "style"),
		}],
		"title": "avoid-get-and-list-prefix",
		"location": {
			"col": 1, "file": "policy.rego", "row": 3,
			"text": `list_users(datasource) := ["we", "have", "no", "users"]`,
		},
		"level": "error",
	}}
}
