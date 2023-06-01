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
			"ref": config.docs.resolve_url("$baseUrl/$category/test-outside-test-package", "testing"),
		}],
		"title": "test-outside-test-package",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": `test_foo { false }`},
		"level": "error",
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

	r := testing.report with input as ast with config.for_rule as {"level": "error"}

	r == {{
		"category": "testing",
		"description": "Files containing tests should have a _test.rego suffix",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/file-missing-test-suffix", "testing"),
		}],
		"title": "file-missing-test-suffix",
		"location": {"file": "policy.rego"},
		"level": "error",
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
			"ref": config.docs.resolve_url("$baseUrl/$category/identically-named-tests", "testing"),
		}],
		"title": "identically-named-tests",
		"location": {"file": "foo_test.rego"},
		"level": "error",
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
			"ref": config.docs.resolve_url("$baseUrl/$category/todo-test", "testing"),
		}],
		"title": "todo-test",
		"location": {"file": "foo_test.rego"},
		"level": "error",
	}}
}

test_fail_call_to_print_and_trace if {
	r := report(`allow {
		print("foo")

		x := [i | i = 0; trace("bar")]
	}`)
	r == {
		{
			"category": "testing",
			"description": "Call to print or trace function",
			"level": "error",
			"location": {"col": 3, "file": "policy.rego", "row": 9, "text": "\t\tprint(\"foo\")"},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/print-or-trace-call", "testing"),
			}],
			"title": "print-or-trace-call",
		},
		{
			"category": "testing",
			"description": "Call to print or trace function",
			"level": "error",
			"location": {"col": 20, "file": "policy.rego", "row": 11, "text": "\t\tx := [i | i = 0; trace(\"bar\")]"},
			"related_resources": [{
				"description": "documentation",
				"ref": config.docs.resolve_url("$baseUrl/$category/print-or-trace-call", "testing"),
			}],
			"title": "print-or-trace-call",
		},
	}
}

report(snippet) := report if {
	# regal ignore:external-reference
	report := testing.report with input as ast.with_future_keywords(snippet)
}
