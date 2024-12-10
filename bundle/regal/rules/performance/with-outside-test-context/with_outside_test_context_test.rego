package regal.rules.performance["with-outside-test-context_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.performance["with-outside-test-context"] as rule

test_fail_with_used_outside_test if {
	module := ast.with_rego_v1(`
	allow if {
		not foo.deny with input as {}
	}
	`)
	r := rule.report with input as module

	r == {{
		"category": "performance",
		"description": "`with` used outside test context",
		"level": "error",
		"location": {
			"col": 16,
			"file": "policy.rego",
			"row": 7,
			"end": {
				"col": 32,
				"row": 7,
			},
			"text": "\t\tnot foo.deny with input as {}",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/with-outside-test-context", "performance"),
		}],
		"title": "with-outside-test-context",
	}}
}

test_success_with_used_in_test if {
	module := ast.with_rego_v1(`
	test_foo_deny if {
		not foo.deny with input as {}
	}
	`)
	r := rule.report with input as module

	r == set()
}

test_success_with_used_in_todo_test if {
	module := ast.with_rego_v1(`
	todo_test_foo_deny if {
		not foo.deny with input as {}
	}
	`)
	r := rule.report with input as module

	r == set()
}
