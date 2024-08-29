package regal.rules.bugs["rule-named-if_test"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.rules.bugs["rule-named-if"] as rule

test_fail_rule_named_if if {
	r := rule.report with input as ast.policy(`
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
