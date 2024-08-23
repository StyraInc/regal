package regal.rules.custom["forbidden-function-call_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.custom["forbidden-function-call"] as rule

test_fail_forbidden_function if {
	module := ast.policy(`foo := http.send({"method": "GET", "url": "https://example.com"})`)

	r := rule.report with input as module with config.for_rule as {
		"level": "error",
		"forbidden-functions": ["http.send"],
	}
		with data.internal.combined_config as {"capabilities": data.regal.capabilities.provided}

	r == {{
		"category": "custom",
		"description": "Forbidden function call",
		"level": "error",
		"location": {
			"col": 8,
			"file": "policy.rego",
			"row": 3,
			"text": `foo := http.send({"method": "GET", "url": "https://example.com"})`,
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/forbidden-function-call", "custom"),
		}],
		"title": "forbidden-function-call",
	}}
}
