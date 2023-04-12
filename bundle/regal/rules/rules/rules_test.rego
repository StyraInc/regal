package regal.rules.rules_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.rules

test_fail_rule_name_shadows_builtin if {
	report(`or := 1`) == {{
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

test_success_rule_name_does_not_shadows_builtin if {
	report(`foo := 1`) == set()
}

report(snippet) := report if {
	# regal ignore:input-or-data-reference
	report := rules.report with input as ast.with_future_keywords(snippet) with config.for_rule as {"enabled": true}
}
