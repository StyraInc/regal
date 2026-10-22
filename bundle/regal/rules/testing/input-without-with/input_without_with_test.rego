package regal.rules.testing["input-without-with_test"]

import data.regal.ast

import data.regal.rules.testing["input-without-with"] as rule

test_fail_input_without_with if {
	module := ast.policy(`
test_something if {
	my_function(input) == "yes"
}
`)

	r := rule.report with input as module

	r == {{
		"category": "testing",
		"description": "Reference to input in test without using with",
		"level": "error",
		"location": {
			"col": 14,
			"end": {
				"col": 19,
				"row": 5,
			},
			"file": "policy.rego",
			"row": 5,
			"text": "\tmy_function(input) == \"yes\"",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/testing/input-without-with",
		}],
		"title": "input-without-with",
	}}
}

test_fail_input_field_without_with if {
	module := ast.policy(`
test_user if {
	input.user == "alice"
}
`)

	r := rule.report with input as module

	r == {{
		"category": "testing",
		"description": "Reference to input in test without using with",
		"level": "error",
		"location": {
			"col": 2,
			"end": {
				"row": 5,
				"col": 12,
			},
			"file": "policy.rego",
			"row": 5,
			"text": "\tinput.user == \"alice\"",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/testing/input-without-with",
		}],
		"title": "input-without-with",
	}}
}

test_success_input_with_with if {
	module := ast.policy(`
test_something if {
	my_function(input) == "yes" with input as {"foo": "bar"}
}
`)

	r := rule.report with input as module

	r == set()
}

test_success_input_field_with_with if {
	module := ast.policy(`
test_user if {
	input.user == "alice" with input.user as "alice"
}
`)

	r := rule.report with input as module

	r == set()
}

test_success_not_a_test_rule if {
	module := ast.policy(`
regular_rule if {
	input.user == "alice"
}
`)

	r := rule.report with input as module

	r == set()
}
