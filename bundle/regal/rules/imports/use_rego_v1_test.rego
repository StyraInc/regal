package regal.rules.imports["use-rego-v1_test"]

import rego.v1

import data.regal.capabilities
import data.regal.config
import data.regal.rules.imports["use-rego-v1"] as rule

test_fail_missing_rego_v1_import if {
	r := rule.report with input as regal.parse_module("policy.rego", `package policy
	import future.keywords

	foo if not bar
	`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == {{
		"category": "imports",
		"description": "Use `import rego.v1`",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-rego-v1", "imports"),
		}],
		"title": "use-rego-v1",
		"location": {"col": 1, "file": "policy.rego", "row": 1, "text": "package policy"},
		"level": "error",
	}}
}

test_success_rego_v1_import if {
	r := rule.report with input as regal.parse_module("policy.rego", `package policy
	import rego.v1

	foo if not bar
	`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}
