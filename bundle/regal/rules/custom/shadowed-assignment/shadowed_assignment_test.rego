package regal.rules.custom["shadowed-assignment_test"]

import data.regal.ast

import data.regal.rules.custom["shadowed-assignment"] as rule

test_shadowed_assignment_rule if {
	module := ast.policy(`
x := 1

r if {
	x := 4
	x == 4
}
`)

	r := rule.report with input as module

	r == {{
		"category": "custom",
		"description": "Shadowed assignment",
		"level": "error",
		"location": {
			"file": "policy.rego",
			"row": 7,
			"col": 2,
			"end": {
				"row": 7,
				"col": 3,
			},
			"text": "\tx := 4",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/custom/shadowed-assignment",
		}],
		"title": "shadowed-assignment",
	}}
}

test_shadowed_assignment_function if {
	module := ast.policy(`
f(a) := a

r if {
	f := 10
}
`)

	r := rule.report with input as module

	r == {{
		"category": "custom",
		"description": "Shadowed assignment",
		"level": "error",
		"location": {
			"file": "policy.rego",
			"row": 7,
			"col": 2,
			"end": {
				"row": 7,
				"col": 3,
			},
			"text": "\tf := 10",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/custom/shadowed-assignment",
		}],
		"title": "shadowed-assignment",
	}}
}

test_shadowed_assignment_import if {
	module := ast.with_rego_v1(`
import data.users

r if {
	users := ["alice", "bob"]
}
`)

	r := rule.report with input as module

	r == {{
		"category": "custom",
		"description": "Shadowed assignment",
		"level": "error",
		"location": {
			"file": "policy.rego",
			"row": 9,
			"col": 2,
			"end": {
				"row": 9,
				"col": 7,
			},
			"text": "\tusers := [\"alice\", \"bob\"]",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/custom/shadowed-assignment",
		}],
		"title": "shadowed-assignment",
	}}
}

test_shadowed_assignment_destructuring if {
	module := ast.policy(`
x := 1

r if {
	[x, y] := [2, 3]
}
`)

	r := rule.report with input as module

	r == {{
		"category": "custom",
		"description": "Shadowed assignment",
		"level": "error",
		"location": {
			"file": "policy.rego",
			"row": 7,
			"col": 3,
			"end": {
				"row": 7,
				"col": 4,
			},
			"text": "\t[x, y] := [2, 3]",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/custom/shadowed-assignment",
		}],
		"title": "shadowed-assignment",
	}}
}

test_not_shadowed_assignment if {
	module := ast.policy(`
x := 1

r if {
	y := 2
	y == 2
}
`)

	r := rule.report with input as module

	r == set()
}
