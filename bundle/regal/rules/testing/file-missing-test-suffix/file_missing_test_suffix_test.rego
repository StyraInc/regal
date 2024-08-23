package regal.rules.testing["file-missing-test-suffix_test"]

import rego.v1

import data.regal.config
import data.regal.rules.testing["file-missing-test-suffix"] as rule

test_fail_test_in_file_without_test_suffix if {
	ast := regal.parse_module("policy.rego", `package foo_test

	test_foo { false }
	`)

	r := rule.report with input as ast with config.for_rule as {"level": "error"}
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
