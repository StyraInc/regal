package regal.rules.rules_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.rules

test_fail_rule_name_shadows_builtin if {
	r := report(`or := 1`)
	r == {{
		"category": "rules",
		"description": "Rule name shadows built-in",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/rule-shadows-builtin",
		}],
		"title": "rule-shadows-builtin",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": "or := 1"},
	}}
}

test_fail_rule_name_starts_with_get if {
	r := report(`get_foo := 1`)
	r == {{
		"category": "rules",
		"description": "Avoid get_ and list_ prefix for rules and functions",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/avoid-get-and-list-prefix",
		}],
		"title": "avoid-get-and-list-prefix",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": "get_foo := 1"},
	}}
}

test_fail_function_name_starts_with_list if {
	r := report(`list_users(datasource) := ["we", "have", "no", "users"]`)
	r == {{
		"category": "rules",
		"description": "Avoid get_ and list_ prefix for rules and functions",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/avoid-get-and-list-prefix",
		}],
		"title": "avoid-get-and-list-prefix",
		"location": {
			"col": 1, "file": "policy.rego", "row": 8,
			"text": `list_users(datasource) := ["we", "have", "no", "users"]`,
		},
	}}
}

test_success_rule_name_does_not_shadows_builtin if {
	report(`foo := 1`) == set()
}

report(snippet) := report if {
	# regal ignore:external-reference
	report := rules.report with input as ast.with_future_keywords(snippet) with config.for_rule as {"enabled": true}
}
