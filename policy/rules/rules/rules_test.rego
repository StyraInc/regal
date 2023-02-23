package regal.rules.rules_test

import future.keywords.if

import data.regal
import data.regal.rules.rules

test_fail_rule_name_shadows_builtin {
	ast := regal.ast(`or := 1`)
	result := rules.violation with input as ast
	result == {{
		"description": "Rule name shadows built-in",
		"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-rules-001"}],
		"title": "STY-RULES-001",
	}}
}

test_success_rule_name_does_not_shadows_builtin {
	ast := regal.ast(`foo := 1`)
	result := rules.violation with input as ast
	result == set()
}
