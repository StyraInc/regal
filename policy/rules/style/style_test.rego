package regal.rules.style_test

import future.keywords.if

import data.regal
import data.regal.rules.style

snake_case_violation := {
	"description": "Prefer snake_case for names",
	"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-style-001"}],
	"title": "STY-STYLE-001",
}

test_fail_camel_cased_rule_name if {
	ast := regal.ast(`camelCase := 5`)
	result := style.violation with input as ast
	result == {snake_case_violation}
}

test_success_snake_cased_rule_name if {
	ast := regal.ast(`snake_case := 5`)
	result := style.violation with input as ast
	result == set()
}

test_fail_camel_cased_some_declaration if {
	ast := regal.ast(`p {some fooBar; input[fooBar]}`)
	result := style.violation with input as ast
	result == {snake_case_violation}
}

test_success_snake_cased_some_declaration if {
	ast := regal.ast(`p {some foo_bar; input[foo_bar]}`)
	result := style.violation with input as ast
	result == set()
}

test_fail_camel_cased_multiple_some_declaration if {
	ast := regal.ast(`p {some x, foo_bar, fooBar; x = 1; foo_bar = 2; input[fooBar]}`)
	result := style.violation with input as ast
	result == {snake_case_violation}
}

test_success_snake_cased_multiple_some_declaration if {
	ast := regal.ast(`p {some x, foo_bar; x = 5; input[foo_bar]}`)
	result := style.violation with input as ast
	result == set()
}

test_fail_camel_cased_var_assignment if {
	ast := regal.ast(`allow { camelCase := 5 }`)
	result := style.violation with input as ast
	result == {snake_case_violation}
}

test_fail_camel_cased_multiple_var_assignment if {
	ast := regal.ast(`allow { snake_case := "foo"; camelCase := 5 }`)
	result := style.violation with input as ast
	result == {snake_case_violation}
}

test_success_snake_cased_var_assignment if {
	ast := regal.ast(`allow { snake_case := 5 }`)
	result := style.violation with input as ast
	result == set()
}

test_fail_camel_cased_some_in_value if {
	ast := regal.ast(`allow { some cC in input }`)
	result := style.violation with input as ast
	result == {snake_case_violation}
}

test_fail_camel_cased_some_in_key_value if {
	ast := regal.ast(`allow { some cC, sc in input }`)
	result := style.violation with input as ast
	result == {snake_case_violation}
}

test_fail_camel_cased_some_in_key_value_2 if {
	ast := regal.ast(`allow { some sc, cC in input }`)
	result := style.violation with input as ast
	result == {snake_case_violation}
}

test_success_snake_cased_some_in if {
	ast := regal.ast(`allow { some sc in input }`)
	result := style.violation with input as ast
	result == set()
}

test_fail_camel_cased_every_value if {
	ast := regal.ast(`allow { every cC in input { cC == 1 } }`)
	result := style.violation with input as ast
	result == {snake_case_violation}
}

test_fail_camel_cased_every_key if {
	ast := regal.ast(`allow { every cC, sc in input { cC == 1; print(sc) } }`)
	result := style.violation with input as ast
	result == {snake_case_violation}
}

test_success_snake_cased_every if {
	ast := regal.ast(`allow { every sc in input { sc == 1 } }`)
	result := style.violation with input as ast
	result == set()
}
