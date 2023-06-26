package regal.rules.testing["test-outside-test-package_test"]

import future.keywords.contains
import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config
import data.regal.rules.testing["test-outside-test-package"] as rule

test_fail_test_outside_test_package if {
	r := rule.report with input as ast.with_future_keywords(`test_foo { false }`)
		with config.for_rule as {"level": "error"}
		with input.regal.file.name as "p_test.rego"

	r == {{
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
	result := rule.report with input as ast
	result == set()
}

# https://github.com/StyraInc/regal/issues/176
test_success_test_prefixed_function if {
	ast := regal.parse_module("foo_test.rego", `
	package foo

	test_foo(x) { x == 1 }
	`)
	result := rule.report with input as ast
	result == set()
}
