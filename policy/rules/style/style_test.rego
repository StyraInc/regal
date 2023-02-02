package regal.rules.style_test

import future.keywords.if

import data.regal
import data.regal.rules.style

test_fail_camel_cased_rule_name if {
	ast := regal.ast(`camelCase := 5`)
	result := style.violation with input as ast
	result == {{
		"description": "Prefer snake_case for names",
		"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-style-001"}],
		"title": "STY-STYLE-001",
	}}
}

test_success_snake_cased_rule_name if {
	ast := regal.ast(`snake_case := 5`)
	result := style.violation with input as ast
	result == set()
}
