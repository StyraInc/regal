package regal.rules.assignment_test

import future.keywords.if

import data.regal
import data.regal.rules.assignment

test_fail_unification_in_default_assignment if {
	report(`default x = false`) == {{
		"category": "assignment",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/use-assignment-operator",
		}],
		"title": "use-assignment-operator",
	}}
}

test_success_assignment_in_default_assignment if {
	report(`default x := false`) == set()
}

test_fail_unification_in_object_rule_assignment if {
	report(`x["a"] = 1`) == {{
		"category": "assignment",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/use-assignment-operator"
		}],
		"title": "assignment-operator"
	}}
}

test_success_assignment_in_object_rule_assignment if {
	report(`x["a"] := 1`) == set()
}

report(snippet) := report {
	report := assignment.report with input as regal.ast(snippet) with regal.rule_config as {"enabled": true}
}

# Blocked by https://github.com/StyraInc/regal/issues/6
#
# allow = true { true }
#
# f(x) = 5
