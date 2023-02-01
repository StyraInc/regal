package regal.rules.assignment_test

import future.keywords.if

import data.regal
import data.regal.rules.assignment

test_fail_unification_in_default_assignment if {
	ast := regal.ast(`default x = false`)
	result := assignment.violation with input as ast
	result == {{
		"description": "Prefer := over = for assignment",
		"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-assign-001"}],
		"scope": "rule",
		"title": "STY-ASSIGN-001",
	}}
}

test_success_assignment_in_default_assignment if {
	ast := regal.ast(`default x := false`)
	result := assignment.violation with input as ast
	result == set()
}

test_fail_unification_in_object_rule_assignment if {
	ast := regal.ast(`x["a"] = 1`)
	result := assignment.violation with input as ast
    print(result)
	result == {{
        "description": "Prefer := over = for assignment",
        "related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-assign-001"}],
        "scope": "rule",
        "title": "STY-ASSIGN-001"
    }}
}

test_success_assignment_in_object_rule_assignment if {
	ast := regal.ast(`x["a"] := 1`)
	result := assignment.violation with input as ast
	result == set()
}
