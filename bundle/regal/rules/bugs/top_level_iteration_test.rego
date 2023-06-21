package regal.rules.bugs_test

import future.keywords.if

import data.regal.config
import data.regal.rules.bugs.common_test.report_with_fk

test_fail_top_level_iteration_wildcard if {
	r := report_with_fk(`x := input.foo.bar[_]`)
	r == {{
		"category": "bugs",
		"description": "Iteration in top-level assignment",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": "x := input.foo.bar[_]"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/top-level-iteration", "bugs"),
		}],
		"title": "top-level-iteration",
		"level": "error",
	}}
}

test_fail_top_level_iteration_named_var if {
	r := report_with_fk(`x := input.foo.bar[i]`)
	r == {{
		"category": "bugs",
		"description": "Iteration in top-level assignment",
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": "x := input.foo.bar[i]"},
		"related_resources": [{
			"description": "documentation",
			"ref": config.docs.resolve_url("$baseUrl/$category/top-level-iteration", "bugs"),
		}],
		"title": "top-level-iteration",
		"level": "error",
	}}
}

test_success_top_level_known_var_ref if {
	report_with_fk(`
	i := "foo"
	x := input.foo.bar[i]`) == set()
}

test_success_top_level_input_ref if {
	report_with_fk(`x := input.foo.bar[input.y]`) == set()
}
