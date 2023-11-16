package regal.rules.bugs["constant-condition_test"]

import rego.v1

import data.regal.ast
import data.regal.config

import data.regal.rules.bugs["constant-condition"] as rule

test_fail_simple_constant_condition if {
	r := rule.report with input as ast.policy(`allow {
	1
	}`)
	r == {{
		"category": "bugs",
		"description": "Constant condition",
		"location": {"col": 2, "file": "policy.rego", "row": 4, "text": "\t1"},
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
	r := rule.report with input as ast.policy(`allow { true }`)
	r == {{
		"category": "bugs",
		"description": "Constant condition",
		"location": {"col": 9, "file": "policy.rego", "row": 3, "text": "allow { true }"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/constant-condition", "bugs"),
		}],
		"title": "constant-condition",
		"level": "error",
	}}
}

test_fail_operator_constant_condition if {
	r := rule.report with input as ast.policy(`allow {
	1 == 1
	}`)
	r == {{
		"category": "bugs",
		"description": "Constant condition",
		"location": {"col": 2, "file": "policy.rego", "row": 4, "text": "\t1 == 1"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/constant-condition", "bugs"),
		}],
		"title": "constant-condition",
		"level": "error",
	}}
}

test_success_non_constant_condition if {
	r := rule.report with input as ast.policy(`allow { 1 == input.one }`)
	r == set()
}

test_success_adding_constant_to_set if {
	r := rule.report with input as ast.with_rego_v1(`rule contains "message"`)
	r == set()
}
