package regal.rules.testnaming.rule2_test

import future.keywords.if
import future.keywords.in

import data.regal.ast
import data.regal.config

import data.regal.rules.testnaming.rule2 as rule

# Example test, replace with your own
test_rule_named_foo_not_allowed {
	module := ast.policy("foo := true")

	r := rule.report with input as module

	# Use print(r) here to see the report. Great for development!

	r == {{
		"category": "testnaming",
		"description": "Add description of rule here!",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "foo := true"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/rule2", "testnaming"),
		}],
		"title": "rule2"
	}}
}
