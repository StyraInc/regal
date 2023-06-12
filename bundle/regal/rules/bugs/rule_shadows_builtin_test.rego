package regal.rules.bugs_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.bugs
import data.regal.rules.bugs.common_test.report

test_fail_rule_name_shadows_builtin if {
	r := report(`or := 1`)
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
	report(`foo := 1`) == set()
}
