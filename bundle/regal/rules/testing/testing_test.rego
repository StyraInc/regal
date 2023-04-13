package regal.rules.testing_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.testing

test_fail_test_outside_test_package if {
	report(`test_foo { false }`) with input.regal.file.name as "p_test.rego" == {{
		"category": "testing",
		"description": "Test outside of test package",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/test-outside-test-package",
		}],
		"title": "test-outside-test-package",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": `test_foo { false }`},
	}}
}

test_success_test_inside_test_package if {
	ast := regal.parse_module("foo_test.rego", `
	package foo_test

	test_foo { false }
	`)
	result := testing.report with input as ast
	result == set()
}

test_fail_test_in_file_without_test_suffix if {
	ast := regal.parse_module("policy.rego", `package foo_test

	test_foo { false }
	`)

	r := testing.report with input as ast with config.for_rule as {"enabled": true}

	r == {{
		"category": "testing",
		"description": "Files containing tests should have a _test.rego suffix",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/file-missing-test-suffix",
		}],
		"title": "file-missing-test-suffix",
		"location": {"file": "policy.rego"},
	}}
}

test_fail_identically_named_tests if {
	ast := regal.parse_module("foo_test.rego", `
	package foo_test

	test_foo { false }
	test_foo { true }
	`)
	result := testing.report with input as ast
	result == {{
		"category": "testing",
		"description": "Multiple tests with same name",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/identically-named-tests",
		}],
		"title": "identically-named-tests",
		"location": {"file": "foo_test.rego"},
	}}
}

test_success_differently_named_tests if {
	ast := regal.parse_module("foo_test.rego", `
	package foo_test

	test_foo { false }
	test_bar { true }
	test_baz { 1 == 1 }
	`)
	result := testing.report with input as ast
	result == set()
}

test_fail_todo_test if {
	ast := regal.parse_module("foo_test.rego", `
	package foo_test

	todo_test_foo { false }

	test_bar { true }
	`)
	result := testing.report with input as ast
	result == {{
		"category": "testing",
		"description": "TODO test encountered",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/todo-test",
		}],
		"title": "todo-test",
		"location": {"file": "foo_test.rego"},
	}}
}

report(snippet) := report if {
	# regal ignore:external-reference
	report := testing.report with input as ast.with_future_keywords(snippet) with config.for_rule as {"enabled": true}
}
