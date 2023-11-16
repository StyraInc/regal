package regal.rules.testing["todo-test_test"]

import rego.v1

import data.regal.config
import data.regal.rules.testing["todo-test"] as rule

test_fail_todo_test if {
	ast := regal.parse_module("foo_test.rego", `
	package foo_test

	todo_test_foo { false }

	test_bar { true }
	`)
	r := rule.report with input as ast
	r == {{
		"category": "testing",
		"description": "TODO test encountered",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/todo-test", "testing"),
		}],
		"title": "todo-test",
		"location": {"col": 2, "file": "foo_test.rego", "row": 4, "text": "\ttodo_test_foo { false }"},
		"level": "error",
	}}
}
