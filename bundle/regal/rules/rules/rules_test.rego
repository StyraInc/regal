package regal.rules.rules_test

import future.keywords.if

import data.regal
import data.regal.rules.rules

test_fail_rule_name_shadows_builtin {
	report(`or := 1`) == {{
		"category": "rules",
		"description": "Rule name shadows built-in",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/rule-shadows-builtin"
		}],
		"title": "rule-shadows-builtin",
	}}
}

test_success_rule_name_does_not_shadows_builtin {
	report(`foo := 1`) == set()
}

report(snippet) := report {
	report := rules.report with input as regal.ast(snippet) with regal.rule_config as {"enabled": true}
}
