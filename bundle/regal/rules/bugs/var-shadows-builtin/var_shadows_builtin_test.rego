package regal.rules.bugs["var-shadows-builtin_test"]

import rego.v1

import data.regal.ast
import data.regal.capabilities
import data.regal.config

import data.regal.rules.bugs["var-shadows-builtin"] as rule

test_fail_var_shadows_builtin if {
	module := ast.with_rego_v1(`allow if http := "yes"`)
	r := rule.report with input as module with data.internal.combined_config as {"capabilities": capabilities.provided}

	r == {{
		"category": "bugs",
		"description": "Variable name shadows built-in",
		"level": "error",
		"location": {
			"col": 10,
			"row": 5,
			"end": {
				"col": 14,
				"row": 5,
			},
			"file": "policy.rego",
			"text": "allow if http := \"yes\"",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/var-shadows-builtin", "bugs"),
		}],
		"title": "var-shadows-builtin",
	}}
}

test_success_var_does_not_shadow_builtin if {
	module := ast.with_rego_v1(`allow if answer := "yes"`)

	r := rule.report with input as module with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}
