package regal.rules.testing["test-outside-test-package_test"]

import data.regal.ast
import data.regal.config
import data.regal.rules.testing["test-outside-test-package"] as rule

test_fail_test_outside_test_package if {
	r := rule.report with input as ast.with_rego_v1(`test_foo if { false }`)
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
		"location": {
			"col": 1,
			"file": "p_test.rego",
			"row": 5,
			"end": {"col": 9, "row": 5},
			"text": `test_foo if { false }`,
		},
		"level": "error",
	}}
}

test_success_test_inside_test_package if {
	ast := regal.parse_module("foo_test.rego", `
	package foo_test

	import rego.v1

	test_foo if { false }
	`)
	result := rule.report with input as ast
	result == set()
}

test_success_test_inside_test_package_named_just_test if {
	ast := regal.parse_module("test.rego", `
	package test

	import rego.v1

	test_foo if { false }
	`)
	result := rule.report with input as ast
	result == set()
}

# https://github.com/StyraInc/regal/issues/176
test_success_test_prefixed_function if {
	ast := regal.parse_module("foo_test.rego", `
	package foo

	import rego.v1

	test_foo(x) if { x == 1 }
	`)
	result := rule.report with input as ast
	result == set()
}
