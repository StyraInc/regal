package regal.rules.testing["identically-named-tests_test"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.config
import data.regal.rules.testing["identically-named-tests"] as rule

test_fail_identically_named_tests if {
	ast := regal.parse_module("foo_test.rego", `
	package foo_test

	test_foo { false }
	test_foo { true }
	`)
	result := rule.report with input as ast
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
	result := rule.report with input as ast
	result == set()
}
