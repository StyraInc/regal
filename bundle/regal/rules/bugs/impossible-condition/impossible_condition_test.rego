package regal.rules.bugs["impossible-condition_test"]

import data.regal.ast

import data.regal.rules.bugs["impossible-condition"] as rule

test_impossible_condition_variable if {
	module := ast.policy(`
allow if {
	foo == "foo"
	foo == "bar"
}
`)

	r := rule.report with input as module

	r == {{
		"category": "bugs",
		"description": "Conflicting conditions in rule body",
		"level": "error",
		"location": {
			"col": 2,
			"end": {
				"col": 14,
				"row": 6,
			},
			"file": "policy.rego",
			"row": 6,
			"text": "\tfoo == \"bar\"",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/bugs/impossible-condition",
		}],
		"title": "impossible-condition",
	}}
}

test_impossible_condition_reference if {
	module := ast.policy(`
allow if {
	input.user == "alice"
	input.user == "bob"
}
`)

	r := rule.report with input as module

	r == {{
		"category": "bugs",
		"description": "Conflicting conditions in rule body",
		"level": "error",
		"location": {
			"col": 2,
			"end": {
				"col": 21,
				"row": 6,
			},
			"file": "policy.rego",
			"row": 6,
			"text": "\tinput.user == \"bob\"",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/bugs/impossible-condition",
		}],
		"title": "impossible-condition",
	}}
}

test_impossible_condition_flipped_operands if {
	module := ast.policy(`
allow if {
	"alice" == input.user
	input.user == "bob"
}
`)

	r := rule.report with input as module

	r == {{
		"category": "bugs",
		"description": "Conflicting conditions in rule body",
		"level": "error",
		"location": {
			"col": 2,
			"end": {
				"col": 21,
				"row": 6,
			},
			"file": "policy.rego",
			"row": 6,
			"text": "\tinput.user == \"bob\"",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/bugs/impossible-condition",
		}],
		"title": "impossible-condition",
	}}
}

test_success_different_targets if {
	module := ast.policy(`
allow if {
	input.user == "alice"
	input.role == "admin"
}
`)

	r := rule.report with input as module

	r == set()
}

test_success_same_condition if {
	module := ast.policy(`
allow if {
	input.user == "alice"
	input.user == "alice"
}
`)

	r := rule.report with input as module

	r == set()
}
