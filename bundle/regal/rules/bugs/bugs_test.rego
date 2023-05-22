package regal.rules.bugs_test

import future.keywords.if

import data.regal.ast
import data.regal.config
import data.regal.rules.bugs

test_fail_simple_constant_condition if {
	r := report(`allow {
	1
	}`)
	r == {{
		"category": "bugs",
		"description": "Constant condition",
		"location": {"col": 2, "file": "policy.rego", "row": 9, "text": "\t1"},
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
		"location": {"col": 2, "file": "policy.rego", "row": 9, "text": "\t1 == 1"},
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

test_fail_top_level_iteration_wildcard if {
	r := report(`x := input.foo.bar[_]`)
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
	r := report(`x := input.foo.bar[i]`)
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
	report(`
	i := "foo"
	x := input.foo.bar[i]`) == set()
}

test_success_top_level_input_ref if {
	report(`x := input.foo.bar[input.y]`) == set()
}

report(snippet) := report if {
	# regal ignore:external-reference
	report := bugs.report with input as ast.with_future_keywords(snippet)
}
