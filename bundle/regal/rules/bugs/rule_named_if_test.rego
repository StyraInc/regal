package regal.rules.bugs_test

import future.keywords.if

import data.regal.config
import data.regal.rules.bugs.common_test.report

test_fail_rule_named_if if {
	r := report(`
	allow := true if {
        input.foo
    }`)
	r == {{
		"category": "bugs",
		"description": "Rule named \"if\"",
		"level": "error",
		"location": {"col": 16, "file": "policy.rego", "row": 4, "text": "\tallow := true if {"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/rule-named-if", "bugs"),
		}],
		"title": "rule-named-if",
	}}
}
