package regal.rules.assignment_test

import future.keywords.if

import data.regal
import data.regal.rules.assignment

test_fail_unification_in_default_assignment if {
	ast := regal.ast("default x = false")
	result := assignment.violation with input as ast
	result == {{
		"description": "Prefer := over = in default assignment",
		"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-unif-001"}],
		"scope": "rule",
		"title": "STY-UNIF-001",
	}}
}

test_success_assignment_in_default_assignment if {
	ast := regal.ast("default x := false")
	result := assignment.violation with input as ast
	result == set()
}
