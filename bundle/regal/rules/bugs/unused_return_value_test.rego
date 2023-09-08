package regal.rules.bugs["unused-return-value_test"]

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.bugs["unused-return-value"] as rule

test_fail_unused_return_value if {
	r := rule.report with input as ast.with_future_keywords(`allow {
		indexof("s", "s")
	}`)
	r == {{
		"category": "bugs",
		"description": "Non-boolean return value unused",
		"level": "error",
		"location": {"col": 3, "file": "policy.rego", "row": 9, "text": "\t\tindexof(\"s\", \"s\")"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/unused-return-value", "bugs"),
		}],
		"title": "unused-return-value",
	}}
}

test_success_unused_boolean_return_value if {
	r := rule.report with input as ast.with_future_keywords(`allow { startswith("s", "s") }`)
	r == set()
}

test_success_return_value_assigned if {
	r := rule.report with input as ast.with_future_keywords(`allow { x := indexof("s", "s") }`)
	r == set()
}

test_success_function_arg_return_ignored if {
	r := rule.report with input as ast.with_future_keywords(`allow {
		indexof("s", "s", i)
	}`)
	r == set()
}
