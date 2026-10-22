package regal.rules.bugs["rule-assignment-conflict_test"]

import data.regal.ast

import data.regal.rules.bugs["rule-assignment-conflict"] as rule

test_fail_rule_assignment_conflict if {
	r := rule.report with input as ast.policy(`
	x := 1
	x := 2
	`)

	r == {{
		"category": "bugs",
		"description": "Rule assignment conflict",
		"level": "error",
		"location": {
			"col": 2,
			"end": {
				"col": 8,
				"row": 5,
			},
			"file": "policy.rego",
			"row": 5,
			"text": "\tx := 2",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/bugs/rule-assignment-conflict",
		}],
		"title": "rule-assignment-conflict",
	}}
}

test_fail_rule_assignment_conflict_assign_and_unassigned if {
	r := rule.report with input as ast.policy(`
	x := 1
	x = 2
	`)

	r == {{
		"category": "bugs",
		"description": "Rule assignment conflict",
		"level": "error",
		"location": {
			"col": 2,
			"end": {
				"col": 7,
				"row": 5,
			},
			"file": "policy.rego",
			"row": 5,
			"text": "\tx = 2",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/bugs/rule-assignment-conflict",
		}],
		"title": "rule-assignment-conflict",
	}}
}

test_fail_rule_assignment_conflict_both_unconditional_equal if {
	r := rule.report with input as ast.policy(`
	x = 1
	x = 2
	`)

	r == {{
		"category": "bugs",
		"description": "Rule assignment conflict",
		"level": "error",
		"location": {
			"col": 2,
			"end": {
				"col": 7,
				"row": 5,
			},
			"file": "policy.rego",
			"row": 5,
			"text": "\tx = 2",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/bugs/rule-assignment-conflict",
		}],
		"title": "rule-assignment-conflict",
	}}
}

test_fail_rule_assignment_conflict_function if {
	r := rule.report with input as ast.policy(`
	f(x) := 1
	f(x) := 2
	`)

	r == {{
		"category": "bugs",
		"description": "Rule assignment conflict",
		"level": "error",
		"location": {
			"col": 2,
			"end": {
				"col": 11,
				"row": 5,
			},
			"file": "policy.rego",
			"row": 5,
			"text": "\tf(x) := 2",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://www.openpolicyagent.org/projects/regal/rules/bugs/rule-assignment-conflict",
		}],
		"title": "rule-assignment-conflict",
	}}
}

test_fail_rule_assignment_conflict_function_wildcard if {
	r := rule.report with input as ast.policy(`
	f(x) := 1
	f(_) := 2
	`)

	_flagged(r) == {"\tf(_) := 2"}
}

test_fail_rule_assignment_conflict_multiple if {
	r := rule.report with input as ast.policy(`
	x := 1
	x := 2
	x := 3
	`)

	_flagged(r) == {"\tx := 2", "\tx := 3"}
}

test_success_conditional_assignments if {
	r := rule.report with input as ast.policy(`
	x := 1 if input.a
	x := 2 if input.b
	`)

	r == set()
}

test_success_default_and_conditional if {
	r := rule.report with input as ast.policy(`
	default x := false

	x := true if input.admin
	`)

	r == set()
}

test_success_boolean_rules if {
	r := rule.report with input as ast.policy(`
	allow if input.a
	allow if input.b
	`)

	r == set()
}

test_success_boolean_rules_with_value if {
	r := rule.report with input as ast.policy(`
	allow = true if input.a
	allow = true if input.b
	`)

	r == set()
}

test_success_multi_value_rules if {
	r := rule.report with input as ast.policy(`
	deny contains "foo" if true
	deny contains "bar" if true
	`)

	r == set()
}

test_success_functions_different_arity if {
	r := rule.report with input as ast.policy(`
	f(x) := 1
	f(x, y) := 2
	`)

	r == set()
}

test_success_dynamic_ref if {
	r := rule.report with input as ast.policy(`
	x[input.y] := 1
	x[input.y] := 2
	`)

	r == set()
}

test_success_function_pattern_matching_args if {
	r := rule.report with input as ast.policy(`
	f("a") := 1
	f("b") := 2
	`)

	r == set()
}

test_success_else_fallback if {
	r := rule.report with input as ast.policy(`
	x := 1 if input.a else := 2
	`)

	r == set()
}

_flagged(report) := {violation.location.text | some violation in report}
