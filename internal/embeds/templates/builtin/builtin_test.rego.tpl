package regal.rules.{{.Category}}{{.NameTest}}

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.{{.Category}}{{.Name}} as rule

# Example test, replace with your own
test_rule_named_foo_not_allowed if {
	module := ast.policy("foo := true")

	r := rule.report with input as module

	# Use print(r) here to see the report. Great for development!

	r == {{ "{{" }}
		"category": "{{.Category}}",
		"description": "Add description of rule here!",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "foo := true"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/{{.NameOriginal}}", "{{.Category}}"),
		}],
		"title": "{{.NameOriginal}}"
	}}
}
