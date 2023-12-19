package regal.rules.bugs["deprecated-builtin_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["deprecated-builtin"] as rule

test_fail_call_to_deprecated_builtin_function if {
	module := ast.with_rego_v1(`
	allow if {
		any([true, false])
	}
	`)

	r := rule.report with input as module with config.capabilities as {"builtins": {"any": {}}}
	r == {{
		"category": "bugs",
		"description": "Avoid using deprecated built-in functions",
		"level": "error",
		"location": {"col": 3, "file": "policy.rego", "row": 7, "text": "\t\tany([true, false])"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/deprecated-builtin", "bugs"),
		}],
		"title": "deprecated-builtin",
	}}
}

test_success_deprecated_builtin_not_in_capabilities if {
	module := ast.with_rego_v1(`
	allow if {
		any([true, false])
	}
	`)

	r := rule.report with input as module with config.capabilities as {"builtins": {"http.send": {}}}
	r == set()
}
