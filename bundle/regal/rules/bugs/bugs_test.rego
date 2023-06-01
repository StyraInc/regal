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

test_fail_unused_return_value if {
	r := report(`allow {
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
	report(`allow { startswith("s", "s") }`) == set()
}

test_success_return_value_assigned if {
	report(`allow { x := indexof("s", "s") }`) == set()
}

test_fail_neq_in_loop if {
	r := report(`deny {
		"admin" != input.user.groups[_]
		input.user.groups[_] != "admin"
	}`)
	r == {
		{
			"category": "bugs",
			"description": "Use of != in loop",
			"level": "error",
			"location": {"col": 11, "file": "policy.rego", "row": 9, "text": "\t\t\"admin\" != input.user.groups[_]"},
			"related_resources": [{
				"description": "documentation",
				"ref": "https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/not-equals-in-loop.md",
			}],
			"title": "not-equals-in-loop",
		},
		{
			"category": "bugs",
			"description": "Use of != in loop",
			"level": "error",
			"location": {"col": 24, "file": "policy.rego", "row": 10, "text": "\t\tinput.user.groups[_] != \"admin\""},
			"related_resources": [{
				"description": "documentation",
				"ref": "https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/not-equals-in-loop.md",
			}],
			"title": "not-equals-in-loop",
		},
	}
}

test_fail_neq_in_loop_one_liner if {
	r := report(`deny if "admin" != input.user.groups[_]`)
	r == {{
		"category": "bugs",
		"description": "Use of != in loop",
		"level": "error",
		"location": {"col": 17, "file": "policy.rego", "row": 8, "text": "deny if \"admin\" != input.user.groups[_]"},
		"related_resources": [{
			"description": "documentation",
			"ref": "https://github.com/StyraInc/regal/blob/main/docs/rules/bugs/not-equals-in-loop.md",
		}],
		"title": "not-equals-in-loop",
	}}
}

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
		"location": {"col": 1, "file": "policy.rego", "row": 8, "text": "or := 1"},
		"level": "error",
	}}
}

test_success_rule_name_does_not_shadows_builtin if {
	report(`foo := 1`) == set()
}

report(snippet) := report if {
	# regal ignore:external-reference
	report := bugs.report with input as ast.with_future_keywords(snippet)
}
