package regal.rules.assignment_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.assignment

test_fail_unification_in_default_assignment if {
	report(`default x = false`) == {{
		"category": "assignment",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-assignment-operator", "assignment"),
		}],
		"title": "use-assignment-operator",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": "default x = false"},
		"level": "error",
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
			"ref": config.docs.resolve_url("$baseUrl/$category/use-assignment-operator", "assignment"),
		}],
		"title": "use-assignment-operator",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": `x["a"] = 1`},
		"level": "error",
	}}
}

test_success_assignment_in_object_rule_assignment if {
	report(`x["a"] := 1`) == set()
}

report(snippet) := report if {
	# regal ignore:external-reference
	report := assignment.report with input as ast.with_future_keywords(snippet)
}

# Blocked by https://github.com/StyraInc/regal/issues/6
#
# allow = true { true }
#
# f(x) = 5
