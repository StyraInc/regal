package regal.rules.bugs_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.bugs.common_test.report

test_fail_simple_constant_condition if {
	r := report(`allow {
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

test_success_static_condition_probably_generated if {
	report(`allow { true }`) == set()
}

test_fail_operator_constant_condition if {
	r := report(`allow {
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
	report(`allow { 1 == input.one }`) == set()
}
