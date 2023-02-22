package regal.rules.testing_test

import future.keywords.if

import data.regal
import data.regal.rules.testing

test_fail_test_outside_test_package if {
	ast := regal.ast(`test_foo { false }`)
	result := testing.violation with input as ast
	result == {{
		"description": "Test outside of test package",
		"related_resources": [{"ref": "https://docs.styra.com/regal/rules/sty-testing-001"}],
		"title": "STY-TESTING-001",
	}}
}

test_success_test_inside_test_package if {
	ast := rego.parse_module("foo_test.rego", `
        package foo_test

        test_foo { false }
    `)
	result := testing.violation with input as ast
	result == set()
}

