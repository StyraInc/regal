package regal.rules.variables_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.variables

test_fail_unconditional_assignment_in_body if {
	report(`x := y { y := 1 }`) == {{
		"category": "variables",
		"description": "Unconditional assignment in rule body",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/unconditional-assignment",
		}],
		"title": "unconditional-assignment",
		"location": {"col": 6, "file": "policy.rego", "row": 8},
	}}
}

test_fail_unconditional_eq_in_body if {
	report(`x = y { y = 1 }`) == {{
		"category": "variables",
		"description": "Unconditional assignment in rule body",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/unconditional-assignment",
		}],
		"title": "unconditional-assignment",
		"location": {"col": 5, "file": "policy.rego", "row": 8},
	}}
}

test_success_conditional_assignment_in_body if {
	report(`x := y { input.foo == "bar"; y := 1 }`) == set()
}

test_success_unconditional_assignment_but_with_in_body if {
	report(`x := y { y := 5 with input as 1 }`) == set()
}

report(snippet) := report if {
	# regal ignore:external-reference
	report := variables.report with input as ast.with_future_keywords(snippet)
		with config.for_rule as {"enabled": true}
}
