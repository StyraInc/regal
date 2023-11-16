package regal.rules.idiomatic["non-raw-regex-pattern_test"]

import rego.v1

import data.regal.ast
import data.regal.capabilities
import data.regal.config

import data.regal.rules.idiomatic["non-raw-regex-pattern"] as rule

test_fail_non_raw_rule_head if {
	r := rule.report with input as ast.policy(`x := regex.match("[0-9]+", "1")`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == {{
		"category": "idiomatic",
		"description": "Use raw strings for regex patterns",
		"level": "error",
		"location": {"col": 18, "file": "policy.rego", "row": 3, "text": "x := regex.match(\"[0-9]+\", \"1\")"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/non-raw-regex-pattern", "idiomatic"),
		}],
		"title": "non-raw-regex-pattern",
	}}
}

test_fail_non_raw_rule_body if {
	r := rule.report with input as ast.policy(`allow {
		regex.is_valid("[0-9]+")
	}`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == {{
		"category": "idiomatic",
		"description": "Use raw strings for regex patterns",
		"level": "error",
		"location": {"col": 18, "file": "policy.rego", "row": 4, "text": "\t\tregex.is_valid(\"[0-9]+\")"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/non-raw-regex-pattern", "idiomatic"),
		}],
		"title": "non-raw-regex-pattern",
	}}
}

test_fail_pattern_in_second_arg if {
	r := rule.report with input as ast.policy(`r := regex.replace("a", "[a]", "b")`)
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == {{
		"category": "idiomatic",
		"description": "Use raw strings for regex patterns",
		"level": "error",
		"location": {"col": 25, "file": "policy.rego", "row": 3, "text": "r := regex.replace(\"a\", \"[a]\", \"b\")"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/non-raw-regex-pattern", "idiomatic"),
		}],
		"title": "non-raw-regex-pattern",
	}}
}

test_success_when_using_raw_string if {
	r := rule.report with input as ast.policy("v := regex.is_valid(`[0-9]+`)")
		with data.internal.combined_config as {"capabilities": capabilities.provided}
	r == set()
}
