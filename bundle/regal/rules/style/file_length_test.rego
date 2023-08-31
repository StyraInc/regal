package regal.rules.style["file-length_test"]

import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config

import data.regal.rules.style["file-length"] as rule

test_fail_configured_file_length_exceeded if {
	module := regal.parse_module("policy.rego", `package policy

	rule1 := "foo"
	rule2 := "bar"
	`)

	r := rule.report with input as module with config.for_rule as {
		"level": "error",
		"max-file-length": 2,
	}
	r == {{
		"category": "style",
		"description": "Max file length exceeded",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 1, "text": "package policy"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/file-length", "style"),
		}],
		"title": "file-length",
	}}
}

test_success_configured_file_length_within_limit if {
	module := regal.parse_module("policy.rego", `package policy

	rule1 := "foo"
	rule2 := "bar"
	`)

	r := rule.report with input as module with config.for_rule as {
		"level": "error",
		"max-file-length": 10,
	}
	r == set()
}
