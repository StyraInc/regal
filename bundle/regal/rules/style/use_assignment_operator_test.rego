package regal.rules.style_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.style
import data.regal.rules.style.common_test.report

test_fail_unification_in_default_assignment if {
	report(`default x = false`) == {{
		"category": "style",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-assignment-operator", "style"),
		}],
		"title": "use-assignment-operator",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": "default x = false"},
		"level": "error",
	}}
}

test_success_assignment_in_default_assignment if {
	report(`default x := false`) == set()
}

test_fail_unification_in_object_rule_assignment if {
	r := report(`x["a"] = 1`)
	r == {{
		"category": "style",
		"description": "Prefer := over = for assignment",
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/use-assignment-operator", "style"),
		}],
		"title": "use-assignment-operator",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": `x["a"] = 1`},
		"level": "error",
	}}
}

test_success_assignment_in_object_rule_assignment if {
	report(`x["a"] := 1`) == set()
}
