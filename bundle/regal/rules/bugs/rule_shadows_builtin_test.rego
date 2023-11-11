package regal.rules.bugs["rule-shadows-builtin_test"]

import rego.v1

import data.regal.ast
import data.regal.config
import data.regal.rules.bugs["rule-shadows-builtin"] as rule

test_fail_rule_name_shadows_builtin if {
	cfg := {"capabilities": {"builtins": {"or": {}}}}
	r := rule.report with input as ast.policy(`or := 1`) with data.internal.combined_config as cfg
	r == {{
		"category": "bugs",
		"description": "Rule name shadows built-in",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/rule-shadows-builtin", "bugs"),
		}],
		"title": "rule-shadows-builtin",
		"location": {"col": 1, "file": "policy.rego", "row": 3, "text": "or := 1"},
		"level": "error",
	}}
}

test_success_rule_name_does_not_shadows_builtin if {
	r := rule.report with input as ast.policy(`foo := 1`)
	r == set()
}
