package regal.rules.variables_test

import future.keywords.if

import data.regal
import data.regal.rules.variables

test_fail_unconditional_assignment_in_body if {
	ast := regal.ast(`x := y { y := 1 }`)
	result := variables.violation with input as ast
	result == {{
		"description": "Unconditional assignment in rule body",
		"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-variables-001"}],
		"title": "STY-VARIABLES-001",
	}}
}

test_fail_unconditional_eq_in_body if {
	ast := regal.ast(`x = y { y = 1 }`)
	result := variables.violation with input as ast
	result == {{
		"description": "Unconditional assignment in rule body",
		"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-variables-001"}],
		"title": "STY-VARIABLES-001",
	}}
}

test_success_conditional_assignment_in_body if {
	ast := regal.ast(`x := y { input.foo == "bar"; y := 1 }`)
	result := variables.violation with input as ast
	result == set()
}
