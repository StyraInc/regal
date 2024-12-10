package regal.rules.bugs["constant-condition_test"]

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["constant-condition"] as rule

test_fail_simple_constant_condition if {
	r := rule.report with input as ast.policy(`allow if {
	1
	}`)
	r == {{
		"category": "bugs",
		"description": "Constant condition",
		"location": {"col": 2, "file": "policy.rego", "row": 4, "text": "\t1", "end": {"row": 4, "col": 3}},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/constant-condition", "bugs"),
		}],
		"title": "constant-condition",
		"level": "error",
	}}
}

test_success_rule_without_body if {
	r := rule.report with input as ast.policy(`allow := true`)
	r == set()
}

test_fail_rule_with_body_looking_generated if {
	r := rule.report with input as ast.policy(`allow if { true }`)
	r == {{
		"category": "bugs",
		"description": "Constant condition",
		"location": {
			"file": "policy.rego",
			"col": 12,
			"row": 3,
			"end": {"row": 3, "col": 16},
			"text": "allow if { true }",
		},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/constant-condition", "bugs"),
		}],
		"title": "constant-condition",
		"level": "error",
	}}
}

test_fail_operator_constant_condition if {
	r := rule.report with input as ast.policy(`allow if {
	1 == 1
	}`)
	r == {{
		"category": "bugs",
		"description": "Constant condition",
		"location": {"col": 2, "file": "policy.rego", "row": 4, "text": "\t1 == 1", "end": {"col": 8, "row": 4}},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/constant-condition", "bugs"),
		}],
		"title": "constant-condition",
		"level": "error",
	}}
}

test_success_non_constant_condition if {
	r := rule.report with input as ast.policy(`allow if { 1 == input.one }`)
	r == set()
}

test_success_adding_constant_to_set if {
	r := rule.report with input as ast.with_rego_v1(`rule contains "message"`)
	r == set()
}
