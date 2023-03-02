package regal.rules.testing_test

import future.keywords.if

import data.regal
import data.regal.rules.testing

test_fail_test_outside_test_package if {
	report(`test_foo { false }`) == {{
		"category": "testing",
		"description": "Test outside of test package",
		"related_resources": [{
			"description": "documentation",
			"ref": "https://docs.styra.com/regal/rules/test-outside-test-package"
		}],
		"title": "test-outside-test-package",
	}}
}

test_success_test_inside_test_package if {
	ast := rego.parse_module("foo_test.rego", `
        package foo_test

        test_foo { false }
    `)
	result := testing.report with input as ast
	result == set()
}

report(snippet) := report {
	report := testing.report with input as regal.ast(snippet) with regal.rule_config as {"enabled": true}
}
