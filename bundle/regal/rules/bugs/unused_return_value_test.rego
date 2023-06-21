package regal.rules.bugs_test

import future.keywords.if

import data.regal.config
import data.regal.rules.bugs.common_test.report_with_fk

test_fail_unused_return_value if {
	r := report_with_fk(`allow {
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
	report_with_fk(`allow { startswith("s", "s") }`) == set()
}

test_success_return_value_assigned if {
	report_with_fk(`allow { x := indexof("s", "s") }`) == set()
}
