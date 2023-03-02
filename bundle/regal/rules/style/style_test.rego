package regal.rules.style_test

import future.keywords.if

import data.regal
import data.regal.rules.style

snake_case_violation := {
	"category": "style",
	"description": "Prefer snake_case for names",
	"related_resources": [{
		"description": "documentation",
		"ref": "https://docs.styra.com/regal/rules/prefer-snake-case"
	}],
	"title": "prefer-snake-case",
}

test_fail_camel_cased_rule_name if {
	report(`camelCase := 5`) == {snake_case_violation}
}

test_success_snake_cased_rule_name if {
	report(`snake_case := 5`) == set()
}

test_fail_camel_cased_some_declaration if {
	report(`p {some fooBar; input[fooBar]}`) == {snake_case_violation}
}

test_success_snake_cased_some_declaration if {
	report(`p {some foo_bar; input[foo_bar]}`) == set()
}

test_fail_camel_cased_multiple_some_declaration if {
	report(`p {some x, foo_bar, fooBar; x = 1; foo_bar = 2; input[fooBar]}`) == {snake_case_violation}
}

test_success_snake_cased_multiple_some_declaration if {
	report(`p {some x, foo_bar; x = 5; input[foo_bar]}`) == set()
}

test_fail_camel_cased_var_assignment if {
	report(`allow { camelCase := 5 }`) == {snake_case_violation}
}

test_fail_camel_cased_multiple_var_assignment if {
	report(`allow { snake_case := "foo"; camelCase := 5 }`) == {snake_case_violation}
}

test_success_snake_cased_var_assignment if {
	report(`allow { snake_case := 5 }`) == set()
}

test_fail_camel_cased_some_in_value if {
	report(`allow { some cC in input }`) == {snake_case_violation}
}

test_fail_camel_cased_some_in_key_value if {
	report(`allow { some cC, sc in input }`) == {snake_case_violation}
}

test_fail_camel_cased_some_in_key_value_2 if {
	report(`allow { some sc, cC in input }`) == {snake_case_violation}
}

test_success_snake_cased_some_in if {
	report(`allow { some sc in input }`) == set()
}

test_fail_camel_cased_every_value if {
	report(`allow { every cC in input { cC == 1 } }`) == {snake_case_violation}
}

test_fail_camel_cased_every_key if {
	report(`allow { every cC, sc in input { cC == 1; print(sc) } }`) == {snake_case_violation}
}

test_success_snake_cased_every if {
	report(`allow { every sc in input { sc == 1 } }`) == set()
}

report(snippet) := report {
	report := style.report with input as regal.ast(snippet) with regal.rule_config as {"enabled": true}
}
