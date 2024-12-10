package regal.rules.bugs["internal-entrypoint_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["internal-entrypoint"] as rule

test_fail_internal_entrypoint if {
	module := ast.with_rego_v1(`

# METADATA
# entrypoint: true
_allow := true
	`)

	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Entrypoint can't be marked internal",
		"level": "error",
		"location": {"col": 1, "file": "policy.rego", "row": 9, "text": "_allow := true", "end": {"col": 7, "row": 9}},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/internal-entrypoint", "bugs"),
		}],
		"title": "internal-entrypoint",
	}}
}

test_fail_internal_entrypoint_rule_ref if {
	module := ast.with_rego_v1(`

# METADATA
# entrypoint: true
authz._allow := true
	`)

	r := rule.report with input as module
	r == {{
		"category": "bugs",
		"description": "Entrypoint can't be marked internal",
		"level": "error",
		"location": {
			"col": 7,
			"end": {"col": 13, "row": 9},
			"file": "policy.rego",
			"row": 9,
			"text": "authz._allow := true",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/internal-entrypoint", "bugs"),
		}],
		"title": "internal-entrypoint",
	}}
}

test_success_non_internal_entrypoint if {
	module := ast.with_rego_v1(`

# METADATA
# entrypoint: true
allow := true
	`)

	r := rule.report with input as module
	r == set()
}
